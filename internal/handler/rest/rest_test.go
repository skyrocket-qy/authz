package rest

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"authz/internal/service"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestHandler_LivenessProbe(t *testing.T) {
	// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
	// pass 'nil' as the third parameter.
	req, err := http.NewRequest("GET", "/healthz/liveness", nil)
	if err != nil {
		t.Fatal(err)
	}

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	handler := NewHandler(nil, nil)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.LivenessProbe(rr, req)

	// Check the status code is what we expect.
	assert.Equal(t, http.StatusOK, rr.Code)

	// Check the response body is what we expect.
	assert.Equal(t, "alive", rr.Body.String())
}

type mockKafkaConnection struct {
	closeErr error
}

func (m *mockKafkaConnection) Close() error {
	return m.closeErr
}

type mockKafkaDialer struct {
	conn service.KafkaConnection
	err  error
}

func (m *mockKafkaDialer) DialLeader(ctx context.Context, network, address string, topic string, partition int) (service.KafkaConnection, error) {
	return m.conn, m.err
}

func TestHandler_ReadinessProbe(t *testing.T) {
	t.Run("should return 200 OK when all checks pass", func(t *testing.T) {
		// Create a mock database
		db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}
		defer db.Close()

		// Expect a ping to the database
		mock.ExpectPing()
		mock.ExpectPing()

		// Create a gorm database
		gormDB, err := gorm.Open(postgres.New(postgres.Config{
			Conn: db,
		}), &gorm.Config{})
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a gorm database connection", err)
		}

		// Create a mock kafka dialer
		kafkaDialer := &mockKafkaDialer{
			conn: &mockKafkaConnection{},
		}

		// Create a request to pass to our handler
		req, err := http.NewRequest("GET", "/healthz/ready", nil)
		if err != nil {
			t.Fatal(err)
		}

		// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
		rr := httptest.NewRecorder()
		handler := NewHandler(gormDB, kafkaDialer)

		// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
		// directly and pass in our Request and ResponseRecorder.
		handler.ReadinessProbe(rr, req)

		// Check the status code is what we expect.
		assert.Equal(t, http.StatusOK, rr.Code)

		// Check the response body is what we expect.
		assert.Equal(t, "ready", rr.Body.String())

		// we make sure that all expectations were met
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})

	t.Run("should return 503 when database ping fails", func(t *testing.T) {
		// Create a mock database
		db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}
		defer db.Close()

		// Expect a ping to the database to fail
		mock.ExpectPing()
		mock.ExpectPing().WillReturnError(fmt.Errorf("db error"))

		// Create a gorm database
		gormDB, err := gorm.Open(postgres.New(postgres.Config{
			Conn: db,
		}), &gorm.Config{})
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a gorm database connection", err)
		}

		// Create a request to pass to our handler
		req, err := http.NewRequest("GET", "/healthz/ready", nil)
		if err != nil {
			t.Fatal(err)
		}

		// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
		rr := httptest.NewRecorder()
		handler := NewHandler(gormDB, nil)

		// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
		// directly and pass in our Request and ResponseRecorder.
		handler.ReadinessProbe(rr, req)

		// Check the status code is what we expect.
		assert.Equal(t, http.StatusServiceUnavailable, rr.Code)

		// Check the response body is what we expect.
		assert.Equal(t, "not ready", rr.Body.String())

		// we make sure that all expectations were met
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})

	t.Run("should return 503 when kafka dial fails", func(t *testing.T) {
		// Create a mock database
		db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}
		defer db.Close()

		// Expect a ping to the database
		mock.ExpectPing()

		// Create a gorm database
		gormDB, err := gorm.Open(postgres.New(postgres.Config{
			Conn: db,
		}), &gorm.Config{})
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a gorm database connection", err)
		}

		// Create a mock kafka dialer that returns an error
		kafkaDialer := &mockKafkaDialer{
			err: fmt.Errorf("kafka error"),
		}

		// Create a request to pass to our handler
		req, err := http.NewRequest("GET", "/healthz/ready", nil)
		if err != nil {
			t.Fatal(err)
		}

		// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
		rr := httptest.NewRecorder()
		handler := NewHandler(gormDB, kafkaDialer)

		// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
		// directly and pass in our Request and ResponseRecorder.
		handler.ReadinessProbe(rr, req)

		// Check the status code is what we expect.
		assert.Equal(t, http.StatusServiceUnavailable, rr.Code)

		// Check the response body is what we expect.
		assert.Equal(t, "not ready", rr.Body.String())

		// we make sure that all expectations were met
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})
}
