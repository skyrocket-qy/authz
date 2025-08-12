package service

import "github.com/segmentio/kafka-go"

func NewKafkaReader() *kafka.Reader {
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{"localhost:9092"},
		Topic:   "test",
	})
}
