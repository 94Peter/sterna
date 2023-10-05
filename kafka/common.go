package kafka

import "context"

type CommonKafka interface {
	GetKafkaWriter() Writer
	Close()
}

func NewCommonKafka(ctx context.Context, di ConfigDI, topic string) CommonKafka {
	return &commonKafkaImpl{
		ctx:      ctx,
		ConfigDI: di,
		topic:    topic,
	}
}

type commonKafkaImpl struct {
	ctx context.Context
	ConfigDI
	kafkaWriter Writer
	topic       string
}

func (s *commonKafkaImpl) GetKafkaWriter() Writer {
	if s.kafkaWriter != nil {
		return s.kafkaWriter
	}
	s.kafkaWriter = s.NewKafkaWriter(s.ctx, s.topic)
	return s.kafkaWriter
}

func (s *commonKafkaImpl) Close() {
	if s.kafkaWriter != nil {
		s.kafkaWriter.Close()
	}
}
