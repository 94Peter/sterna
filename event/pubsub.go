package event

import (
	"context"
	"fmt"
	"sync"

	"cloud.google.com/go/pubsub"
	"cloud.google.com/go/pubsublite/pscompat"
	"github.com/94peter/sterna/log"
	"golang.org/x/sync/errgroup"
	"google.golang.org/api/option"
)

func NewPubSubEvent() Event {
	return &pubsubEvent{
		ctx: context.Background(),
	}
}

type pubsubEvent struct {
	ctx             context.Context
	credentialsFile string
	projectID       string
	zone            string
	log             log.Logger
}

func (p *pubsubEvent) Close() {
}

func (p *pubsubEvent) Fire(topic string, datas [][]byte) error {
	// projectID := "517745428365"
	// zone := "asia-east1-a"
	// topicID := "donation"

	messageCount := len(datas)
	topicPath := fmt.Sprintf("projects/%s/locations/%s/topics/%s", p.projectID, p.zone, topic)

	// Create the publisher client.
	publisher, err := pscompat.NewPublisherClient(p.ctx, topicPath, option.WithCredentialsFile(p.credentialsFile))

	if err != nil {
		p.log.Fatal(fmt.Sprintf("pscompat.NewPublisherClient error: %v", err))
		return err
	}

	// Ensure the publisher will be shut down.
	defer publisher.Stop()

	// Collect any messages that need to be republished with a new publisher
	// client.
	var toRepublish []*pubsub.Message
	var toRepublishMu sync.Mutex

	// Publish messages. Messages are automatically batched.
	g := new(errgroup.Group)
	for i := 0; i < messageCount; i++ {
		msg := &pubsub.Message{
			Data: datas[i],
		}
		result := publisher.Publish(p.ctx, msg)

		g.Go(func() error {
			// Get blocks until the result is ready.
			id, err := result.Get(p.ctx)
			if err != nil {
				// NOTE: A failed PublishResult indicates that the publisher client
				// encountered a fatal error and has permanently terminated. After the
				// fatal error has been resolved, a new publisher client instance must
				// be created to republish failed messages.
				p.log.Err(fmt.Sprintf("Publish error: %v", err))
				toRepublishMu.Lock()
				toRepublish = append(toRepublish, msg)
				toRepublishMu.Unlock()
				return err
			}

			// Metadata decoded from the id contains the partition and offset.
			metadata, err := pscompat.ParseMessageMetadata(id)
			if err != nil {
				p.log.Err(fmt.Sprintf("Failed to parse message metadata %q: %v", id, err))
				return err
			}
			p.log.Err(fmt.Sprintf("Published: partition=%d, offset=%d", metadata.Partition, metadata.Offset))
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		p.log.Err(fmt.Sprintf("Publishing finished with error: %v", err))
	}

	p.log.Info(fmt.Sprintf("Published %d messages", messageCount-len(toRepublish)))
	// Print the error that caused the publisher client to terminate (if any),
	// which may contain more context than PublishResults.
	if err := publisher.Error(); err != nil {
		p.log.Err(fmt.Sprintf("Publisher client terminated due to error: %v\n", publisher.Error()))
		return err
	}
	return nil
}

func (p *pubsubEvent) Register(jobs ...EventJob) {
	// projectID := "517745428365"
	// zone := "asia-east1-a"
	// subscriptionID := "test"

	// flag.Parse()

	settings := pscompat.ReceiveSettings{
		// 10 MiB. Must be greater than the allowed size of the largest message
		// (1 MiB).
		MaxOutstandingBytes: 10 * 1024 * 1024,
		// 1,000 outstanding messages. Must be > 0.
		MaxOutstandingMessages: 1000,
	}
	for _, j := range jobs {
		go func(t string, h func(data []byte) error) {

			subscriptionPath := fmt.Sprintf("projects/%s/locations/%s/subscriptions/%s", p.projectID, p.zone, t)
			subscriber, err := pscompat.NewSubscriberClientWithSettings(p.ctx, subscriptionPath, settings, option.WithCredentialsFile(p.credentialsFile))
			if err != nil {
				p.log.Fatal(fmt.Sprintf("pscompat.NewSubscriberClientWithSettings error: %v", err))
			}
			p.log.Info(fmt.Sprintf("Listening to messages on %s...\n", subscriptionPath))
			if err := subscriber.Receive(p.ctx, func(ctx context.Context, msg *pubsub.Message) {
				// NOTE: May be called concurrently; synchronize access to shared memory.

				// Metadata decoded from the message ID contains the partition and offset.
				metadata, err := pscompat.ParseMessageMetadata(msg.ID)
				if err != nil {
					p.log.Fatal(fmt.Sprintf("Failed to parse %q: %v", msg.ID, err))
				}
				err = h(msg.Data)
				if err != nil {
					p.log.Err(fmt.Sprintf("data handler err: %v", err))
				}
				p.log.Debug(fmt.Sprintf("Received (partition=%d, offset=%d): %s\n", metadata.Partition, metadata.Offset, string(msg.Data)))
				msg.Ack()
			}); err != nil {
				p.log.Fatal(fmt.Sprintf("SubscriberClient.Receive error: %v", err))
			}
		}(j.GetTopic(), j.GetHandler())

	}
}
