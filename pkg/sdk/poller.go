package sdk

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// Poller polls the config server for updates
type Poller struct {
	client      *Client
	environment string
	interval    time.Duration
	stopCh      chan struct{}
	running     bool
}

// NewPoller creates a new Poller instance
func NewPoller(client *Client, environment string, interval time.Duration) *Poller {
	return &Poller{
		client:      client,
		environment: environment,
		interval:    interval,
		stopCh:      make(chan struct{}),
	}
}

// Start begins polling for config updates
func (p *Poller) Start(ctx context.Context) {
	if p.running {
		return
	}
	p.running = true

	go func() {
		ticker := time.NewTicker(p.interval)
		defer ticker.Stop()

		// Initial fetch
		p.poll(ctx)

		for {
			select {
			case <-ticker.C:
				p.poll(ctx)
			case <-p.stopCh:
				return
			case <-ctx.Done():
				return
			}
		}
	}()
}

// Stop stops the poller
func (p *Poller) Stop() {
	if p.running {
		close(p.stopCh)
		p.running = false
	}
}

// poll fetches all configs and updates the local store
func (p *Poller) poll(ctx context.Context) {
	url := fmt.Sprintf("%s/api/v1/configs/%s", p.client.baseURL, p.environment)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		log.Printf("config-sdk: failed to create poll request: %v", err)
		return
	}

	req.Header.Set("X-API-Key", p.client.apiKey)

	resp, err := p.client.httpClient.Do(req)
	if err != nil {
		log.Printf("config-sdk: poll request failed: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("config-sdk: poll request returned status %d", resp.StatusCode)
		return
	}

	var result struct {
		Configs []struct {
			Key       string      `json:"key"`
			Value     interface{} `json:"value"`
			ValueType string      `json:"value_type"`
		} `json:"configs"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Printf("config-sdk: failed to decode poll response: %v", err)
		return
	}

	// Update local store
	values := make(map[string]ConfigValue)
	for _, cfg := range result.Configs {
		values[cfg.Key] = ConfigValue{
			Value:     cfg.Value,
			ValueType: cfg.ValueType,
		}
	}
	p.client.store.SetAll(values)
}
