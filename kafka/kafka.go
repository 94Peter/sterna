package kafka

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/94peter/sterna/log"
	"github.com/segmentio/kafka-go"
)

type ConfigDI interface {
	NewKafkaWriter(ctx context.Context, topic string) Writer
	NewKafkaReader(ctx context.Context, groupID, topic string, l log.Logger) Reader
}

type KafkaConfig struct {
	Brokers []string
}

func (c *KafkaConfig) NewKafkaWriter(ctx context.Context, topic string) Writer {
	return &writerImpl{
		ctx: ctx,
		kafka: &kafka.Writer{
			Addr:     kafka.TCP(c.Brokers...),
			Topic:    topic,
			Balancer: &kafka.LeastBytes{},
		},
	}
}

type Writer interface {
	SetLog(log.Logger)
	Message(headers map[string][]byte, msg []byte) error
	Close() error
}

type writerImpl struct {
	ctx   context.Context
	kafka *kafka.Writer
	l     log.Logger
}

func (wi *writerImpl) SetLog(l log.Logger) {
	wi.l = l
}

func (wi *writerImpl) Message(headers map[string][]byte, msg []byte) error {
	var myheaders []kafka.Header
	for k, v := range headers {
		myheaders = append(myheaders, kafka.Header{
			Key:   k,
			Value: v,
		})
	}
	m := kafka.Message{
		Value:   msg,
		Headers: myheaders,
	}
	if wi.l != nil {
		wi.l.Info(fmt.Sprintf(
			"write kafka topic [%s], header [%v] message: %s",
			wi.kafka.Topic, myheaders, string(msg)))
	}
	err := wi.kafka.WriteMessages(wi.ctx, m)
	if err != nil {
		return err
	}
	return nil
}

func (wi *writerImpl) Close() error {
	return wi.kafka.Close()
}

type Reader interface {
	Read() (map[string]string, []byte, error)
	ReadHandler(handler ReaderHandler) error
	Close() error
}

func (c *KafkaConfig) NewKafkaReader(ctx context.Context, groupID, topic string, l log.Logger) Reader {
	dial := kafka.DefaultDialer
	hostname, _ := os.Hostname()
	dial.ClientID = "kafka@" + hostname
	return &readerImpl{
		ctx: ctx,
		kafka: kafka.NewReader(kafka.ReaderConfig{
			Dialer:   dial,
			Brokers:  c.Brokers,
			GroupID:  groupID,
			Topic:    topic,
			MinBytes: 10e3, // 10KB
			MaxBytes: 10e6, // 10MB
		}),
		log: l,
	}
}

type readerImpl struct {
	ctx   context.Context
	kafka *kafka.Reader
	log   log.Logger
}

type ReaderHandler func(headers map[string]string, data []byte) error

func (ri *readerImpl) Read() (map[string]string, []byte, error) {
	m, err := ri.kafka.ReadMessage(ri.ctx)
	if err != nil {
		return nil, nil, err
	}
	if ri.log != nil {
		ri.log.Debug(fmt.Sprintf("message at topic: %v partition: %v offset: %v value: %s\n", m.Topic, m.Partition, m.Offset, string(m.Value)))
	}
	var headers map[string]string
	if len(m.Headers) > 0 {
		headers = map[string]string{}
		for _, h := range m.Headers {
			headers[h.Key] = string(h.Value)
		}
	}
	return headers, m.Value, nil
}

const waitingTime = time.Second * 5
const maxRetriedTimes = 5

func (ri *readerImpl) ReadHandler(handler ReaderHandler) error {
	m, err := ri.kafka.FetchMessage(ri.ctx)
	if err != nil {
		return errors.New("fetch message fail: " + err.Error())
	}
	var headers map[string]string
	if len(m.Headers) > 0 {
		headers = map[string]string{}
		for _, h := range m.Headers {
			headers[h.Key] = string(h.Value)
		}
	}
	retriedTimes := 0
	for err = handler(headers, m.Value); err != nil; err = handler(headers, m.Value) {
		ri.log.Warn("halder data fail: " + err.Error())
		ri.log.Info("waiting for 5 secs to retry")
		retriedTimes++
		if retriedTimes >= maxRetriedTimes {
			return err
		}
		time.Sleep(waitingTime)
	}

	if err = ri.kafka.CommitMessages(ri.ctx, m); err != nil {
		return errors.New("commit messages fail: " + err.Error())
	}

	return nil
}

func (ri *readerImpl) Close() error {
	return ri.kafka.Close()
}
