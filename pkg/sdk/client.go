package sdk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

// Client is the config-manager SDK client
type Client struct {
	baseURL       string
	apiKey        string
	environmentID string
	httpClient    *http.Client
	store         *Store
	poller        *Poller
}

// Config holds client configuration options
type Config struct {
	BaseURL       string
	APIKey        string
	EnvironmentID string        // Environment ID to fetch configs from
	PollInterval  time.Duration // How often to poll for updates (0 = no polling)
	HTTPTimeout   time.Duration // HTTP request timeout (default: 10s)
}

// ConfigValue represents a typed configuration value
type ConfigValue struct {
	Value     interface{} `json:"value"`
	ValueType string      `json:"value_type"`
}

// NewClient creates a new config-manager client
func NewClient(cfg Config) *Client {
	timeout := cfg.HTTPTimeout
	if timeout == 0 {
		timeout = 10 * time.Second
	}

	client := &Client{
		baseURL:       cfg.BaseURL,
		apiKey:        cfg.APIKey,
		environmentID: cfg.EnvironmentID,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		store: NewStore(),
	}

	if cfg.PollInterval > 0 {
		client.poller = NewPoller(client, cfg.EnvironmentID, cfg.PollInterval)
	}

	return client
}

// Get retrieves a config value by key (returns interface{})
func (c *Client) Get(ctx context.Context, key string) (interface{}, error) {
	// First check local store
	if value, ok := c.store.Get(key); ok {
		return value.Value, nil
	}

	// Fetch from server
	return c.fetchConfig(ctx, key)
}

// GetString retrieves a config value as string
func (c *Client) GetString(ctx context.Context, key string) (string, error) {
	value, err := c.Get(ctx, key)
	if err != nil {
		return "", err
	}
	if str, ok := value.(string); ok {
		return str, nil
	}
	return fmt.Sprintf("%v", value), nil
}

// GetInt retrieves a config value as int
func (c *Client) GetInt(ctx context.Context, key string) (int, error) {
	value, err := c.Get(ctx, key)
	if err != nil {
		return 0, err
	}

	switch v := value.(type) {
	case int:
		return v, nil
	case float64: // JSON numbers are float64
		return int(v), nil
	case string:
		return strconv.Atoi(v)
	default:
		return 0, fmt.Errorf("cannot convert %T to int", value)
	}
}

// GetBool retrieves a config value as bool
func (c *Client) GetBool(ctx context.Context, key string) (bool, error) {
	value, err := c.Get(ctx, key)
	if err != nil {
		return false, err
	}

	switch v := value.(type) {
	case bool:
		return v, nil
	case string:
		return strconv.ParseBool(v)
	default:
		return false, fmt.Errorf("cannot convert %T to bool", value)
	}
}

// GetFloat retrieves a config value as float64
func (c *Client) GetFloat(ctx context.Context, key string) (float64, error) {
	value, err := c.Get(ctx, key)
	if err != nil {
		return 0, err
	}

	switch v := value.(type) {
	case float64:
		return v, nil
	case int:
		return float64(v), nil
	case string:
		return strconv.ParseFloat(v, 64)
	default:
		return 0, fmt.Errorf("cannot convert %T to float64", value)
	}
}

// GetJSON retrieves a config value as map/slice
func (c *Client) GetJSON(ctx context.Context, key string) (interface{}, error) {
	return c.Get(ctx, key)
}

// GetWithDefault retrieves a config value or returns the default
func (c *Client) GetWithDefault(ctx context.Context, key string, defaultValue interface{}) interface{} {
	value, err := c.Get(ctx, key)
	if err != nil {
		return defaultValue
	}
	return value
}

// GetStringWithDefault retrieves a string config or returns default
func (c *Client) GetStringWithDefault(ctx context.Context, key, defaultValue string) string {
	value, err := c.GetString(ctx, key)
	if err != nil {
		return defaultValue
	}
	return value
}

// GetIntWithDefault retrieves an int config or returns default
func (c *Client) GetIntWithDefault(ctx context.Context, key string, defaultValue int) int {
	value, err := c.GetInt(ctx, key)
	if err != nil {
		return defaultValue
	}
	return value
}

// GetBoolWithDefault retrieves a bool config or returns default
func (c *Client) GetBoolWithDefault(ctx context.Context, key string, defaultValue bool) bool {
	value, err := c.GetBool(ctx, key)
	if err != nil {
		return defaultValue
	}
	return value
}

// Set creates or updates a config value
func (c *Client) Set(ctx context.Context, key string, value interface{}, performedBy string) error {
	url := fmt.Sprintf("%s/api/v1/configs/%s/%s", c.baseURL, c.environmentID, key)

	reqBody := map[string]interface{}{
		"value":        value,
		"performed_by": performedBy,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return err
	}

	req.Header.Set("X-API-Key", c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to set config: status %d", resp.StatusCode)
	}

	// Update local cache
	var result struct {
		Value     interface{} `json:"value"`
		ValueType string      `json:"value_type"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	c.store.Set(key, ConfigValue{Value: result.Value, ValueType: result.ValueType})

	return nil
}

// Delete removes a config value
func (c *Client) Delete(ctx context.Context, key, performedBy string) error {
	url := fmt.Sprintf("%s/api/v1/configs/%s/%s", c.baseURL, c.environmentID, key)

	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("X-API-Key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("failed to delete config: status %d", resp.StatusCode)
	}

	// Remove from local cache
	c.store.Delete(key)

	return nil
}

// ListAll retrieves all configs for the environment
func (c *Client) ListAll(ctx context.Context) (map[string]ConfigValue, error) {
	url := fmt.Sprintf("%s/api/v1/configs/%s", c.baseURL, c.environmentID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-API-Key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to list configs: status %d", resp.StatusCode)
	}

	var result struct {
		Configs []struct {
			Key       string      `json:"key"`
			Value     interface{} `json:"value"`
			ValueType string      `json:"value_type"`
		} `json:"configs"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	configs := make(map[string]ConfigValue)
	for _, cfg := range result.Configs {
		configs[cfg.Key] = ConfigValue{
			Value:     cfg.Value,
			ValueType: cfg.ValueType,
		}
	}

	return configs, nil
}

// fetchConfig fetches a config from the server
func (c *Client) fetchConfig(ctx context.Context, key string) (interface{}, error) {
	url := fmt.Sprintf("%s/api/v1/configs/%s/%s", c.baseURL, c.environmentID, key)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-API-Key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get config: status %d", resp.StatusCode)
	}

	var result struct {
		Value     interface{} `json:"value"`
		ValueType string      `json:"value_type"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	// Cache in local store
	c.store.Set(key, ConfigValue{Value: result.Value, ValueType: result.ValueType})

	return result.Value, nil
}

// StartPolling starts background polling for config updates
func (c *Client) StartPolling(ctx context.Context) {
	if c.poller != nil {
		c.poller.Start(ctx)
	}
}

// StopPolling stops background polling
func (c *Client) StopPolling() {
	if c.poller != nil {
		c.poller.Stop()
	}
}

// Close cleans up client resources
func (c *Client) Close() {
	c.StopPolling()
}
