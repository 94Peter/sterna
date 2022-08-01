package kafka

import (
	"context"
	"fmt"
	"os"

	"github.com/94peter/sterna/log"
	"github.com/segmentio/kafka-go"
)

type ConfigDI interface {
	NewWriter(ctx context.Context, topic string) Writer
	NewReader(ctx context.Context, groupID, topic string, l log.Logger) Reader
}

type Config struct {
	Brokers []string
}

func (c *Config) NewWriter(ctx context.Context, topic string) Writer {
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
	Message(msg []byte) error
	Close() error
}

type writerImpl struct {
	ctx   context.Context
	kafka *kafka.Writer
}

func (wi *writerImpl) Message(msg []byte) error {
	m := kafka.Message{
		Value: msg,
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
	Read() ([]byte, error)
	Close() error
}

func (c *Config) NewReader(ctx context.Context, groupID, topic string, l log.Logger) Reader {
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

func (ri *readerImpl) Read() ([]byte, error) {
	m, err := ri.kafka.ReadMessage(ri.ctx)
	if err != nil {
		return nil, err
	}
	if ri.log != nil {
		ri.log.Debug(fmt.Sprintf("message at topic: %v partition: %v offset: %v value: %s\n", m.Topic, m.Partition, m.Offset, string(m.Value)))
	}
	return m.Value, nil
}

func (ri *readerImpl) Close() error {
	return ri.kafka.Close()
}
