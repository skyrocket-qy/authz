package main

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"
)

type LokiPayload struct {
	Streams []Stream `json:"streams"`
}

type Stream struct {
	Stream map[string]string `json:"stream"`
	Values [][]string        `json:"values"`
}

func sendLogToLoki(level, message string) {
	// Construct the payload
	timestamp := strconv.FormatInt(time.Now().UnixNano(), 10)
	payload := LokiPayload{
		Streams: []Stream{
			{
				Stream: map[string]string{
					"job":   "my-go-server",
					"env":   "production",
					"level": level,
				},
				Values: [][]string{
					{timestamp, message},
				},
			},
		},
	}

	// Convert to JSON
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Failed to marshal payload: %v", err)

		return
	}

	// Send to Loki
	req, err := http.NewRequestWithContext(
		context.TODO(),
		http.MethodPost,
		"http://localhost:3100/loki/api/v1/push",
		bytes.NewBuffer(payloadBytes),
	)
	if err != nil {
		log.Printf("Failed to create request to Loki: %v", err)

		return
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Failed to send log to Loki: %v", err)

		return
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Print("Failed to close response body")
		}
	}()

	// Check response
	if resp.StatusCode != http.StatusNoContent {
		log.Printf("Unexpected Loki response: %d %s", resp.StatusCode, resp.Status)
	}
}

func main() {
	sendLogToLoki("info", "This is an info log sent to Loki")
	sendLogToLoki("error", "This is an error log sent to Loki")
	sendLogToLoki("warning", "This is an warning log sent to Loki")
	sendLogToLoki("debug", "This is an debug log sent to Loki")
	sendLogToLoki("panic", "This is an panic log sent to Loki")
	sendLogToLoki("fatal", "This is an fatal log sent to Loki")
	sendLogToLoki("trace", "This is an trace log sent to Loki")
}
