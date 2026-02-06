package logging

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type lokiPushPayload struct {
	Streams []lokiStream `json:"streams"`
}

type lokiStream struct {
	Stream map[string]string `json:"stream"`
	Values [][]string        `json:"values"`
}

func SendToLoki(serviceName, logLevel, message string) error {
	lokiURL := os.Getenv("LOKI_PUSH_URL")
	if lokiURL == "" {
		lokiURL = "http://localhost:3100/loki/api/v1/push"
	}

	timestamp := fmt.Sprintf("%d", time.Now().UnixNano())
	payload := lokiPushPayload{
		Streams: []lokiStream{
			{
				Stream: map[string]string{
					"service": serviceName,
					"level":   logLevel,
				},
				Values: [][]string{{timestamp, message}},
			},
		},
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, lokiURL, bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("loki push failed: %s", string(body))
	}

	return nil
}
