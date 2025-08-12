package service

import "github.com/segmentio/kafka-go"

func NewKafkaReader() *kafka.Reader {
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{"localhost:29092"},
		Topic:   "pg.public.tuples",
	})
}
