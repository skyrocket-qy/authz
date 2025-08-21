package service

import (
	"fmt"

	"authz/internal/cfg"
	"authz/internal/pkg"
	"github.com/rs/zerolog/log"
	"github.com/segmentio/kafka-go"
)

func NewKafkaReader(lc *pkg.LifecycleParallel) *kafka.Reader {
	log.Info().Msgf("start kafka reader")

	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{fmt.Sprintf("%s:%s", cfg.Cfg.Kafka.Host, cfg.Cfg.Kafka.Port)},
		Topic:   "pg.public.tuples",
	})

	lc.Add(r, r.Close)

	return r
}
