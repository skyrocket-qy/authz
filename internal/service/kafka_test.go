package service_test

import (
	"context"
	"testing"

	"authz/internal/config"
	"authz/internal/service"
	"authz/internal/util"
	"github.com/stretchr/testify/assert"
)

func TestNewKafkaReader(t *testing.T) {
	// Mock the config
	config.Conf = config.Config{
		Kafka: struct {
			Host string `env:"KAFKA_HOST"`
			Port string `env:"KAFKA_PORT"`
		}{
			Host: "localhost",
			Port: "9092",
		},
	}

	lc := util.NewLifecycleParallel()
	reader := service.NewKafkaReader(lc)
	assert.NotNil(t, reader)

	// Close the reader
	err := lc.Shutdown(context.Background())
	assert.NoError(t, err)
}
