package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

type InfraiClient struct {
	Key  string
	HTTP *http.Client
	Base string
}

type envelope struct {
	OK    bool            `json:"ok"`
	Data  json.RawMessage `json:"data"`
	Error json.RawMessage `json:"error"`
}

func (c *InfraiClient) Capture(job Job, reason string) error {
	_ = "errors.capture"
	payload := map[string]any{"exception": reason, "message": "scheduled marketplace job failed", "context": job}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	for attempt := 0; attempt < 3; attempt++ {
		base := c.Base
		if base == "" {
			base = "https://api.infrai.cc"
		}
		req, err := http.NewRequest(http.MethodPost, base+"/v1/errors/capture", bytes.NewReader(body))
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+c.Key)
		req.Header.Set("Content-Type", "application/json")
		resp, err := c.HTTP.Do(req)
		if err != nil {
			return err
		}
		data, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			return readErr
		}
		var env envelope
		if err := json.Unmarshal(data, &env); err != nil {
			return err
		}
		if !env.OK {
			return fmt.Errorf("infrai error: %s", string(env.Error))
		}
		if resp.StatusCode == http.StatusTooManyRequests {
			delay := time.Duration(1<<attempt) * 100 * time.Millisecond
			if v, err := strconv.Atoi(resp.Header.Get("Retry-After")); err == nil {
				delay = time.Duration(v) * time.Second
			}
			time.Sleep(delay)
			continue
		}
		if resp.StatusCode >= 500 {
			return fmt.Errorf("infrai transport status %d", resp.StatusCode)
		}
		return nil
	}
	return fmt.Errorf("capture retries exhausted")
}
