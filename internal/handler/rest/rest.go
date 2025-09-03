package rest

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"authz/internal/config"
	"authz/internal/service"
	"gorm.io/gorm"
)

type Handler struct {
	pgdb        *gorm.DB
	kafkaDialer service.KafkaDialer
}

func NewHandler(pgdb *gorm.DB, kafkaDialer service.KafkaDialer) *Handler {
	return &Handler{
		pgdb:        pgdb,
		kafkaDialer: kafkaDialer,
	}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/healthz/liveness", h.LivenessProbe)
	mux.HandleFunc("/healthz/ready", h.ReadinessProbe)
}

// LivenessProbe handles GET /healthz/liveness
// Only checks that the app process hasn’t crashed/hung.
func (h *Handler) LivenessProbe(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("alive"))
}

// ReadinessProbe handles GET /healthz/ready
// Checks that the app is ready to handle requests, e.g., DB connected.
func (h *Handler) ReadinessProbe(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// 1️⃣ Check Postgres
	sqlDB, err := h.pgdb.DB()
	if err != nil || sqlDB.PingContext(ctx) != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte("not ready"))

		return
	}

	// 2️⃣ Check Kafka (connection only, don't consume messages)
	conn, err := h.kafkaDialer.DialLeader(ctx, "tcp",
		fmt.Sprintf("%s:%s", config.Conf.Kafka.Host, config.Conf.Kafka.Port), "pg.public.tuples", 0,
	)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte("not ready"))

		return
	}

	_ = conn.Close()

	// ✅ All checks passed
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ready"))
}
