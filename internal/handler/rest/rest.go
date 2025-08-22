package rest

import (
	"authz/internal/cfg"
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/segmentio/kafka-go"
	"gorm.io/gorm"
)

type Handler struct {
	pgdb *gorm.DB
}

func NewHandler(pgdb *gorm.DB) *Handler {
	return &Handler{
		pgdb: pgdb,
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
	w.Write([]byte("alive"))
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
		w.Write([]byte("not ready"))
		return
	}

	// 2️⃣ Check Kafka (connection only, don't consume messages)
	conn, err := kafka.DialLeader(ctx, "tcp",
		fmt.Sprintf("%s:%s", cfg.Cfg.Kafka.Host, cfg.Cfg.Kafka.Port), "pg.public.tuples", 0,
	)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("not ready"))
		return
	}
	conn.Close()

	// ✅ All checks passed
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ready"))
}
