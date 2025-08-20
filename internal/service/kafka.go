package service

import (
	"authz/internal/pkg"
	"github.com/segmentio/kafka-go"
)

func NewKafkaReader(lc *pkg.LifecycleParallel) *kafka.Reader {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{"localhost:29092"},
		Topic:   "pg.public.tuples",
	})

	lc.Add(r, r.Close)

	return r
}
