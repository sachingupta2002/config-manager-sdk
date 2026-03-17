package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/sachin/config-manager/pkg/sdk"
)

func main() {
	// Initialize the SDK client
	client := sdk.NewClient(sdk.Config{
		BaseURL:       "http://localhost:8080",
		APIKey:        "your-api-key-here",
		EnvironmentID: "your-environment-id", // Get from /api/v1/environments
		PollInterval:  30 * time.Second,      // Poll for updates every 30s
		HTTPTimeout:   10 * time.Second,
	})
	defer client.Close()

	ctx := context.Background()

	// Start background polling (optional)
	client.StartPolling(ctx)

	// Example 1: Get typed values
	fmt.Println("=== Reading Configs ===")

	// Get string value
	dbHost, err := client.GetString(ctx, "database.host")
	if err != nil {
		log.Printf("Error getting database.host: %v", err)
	} else {
		fmt.Printf("Database Host: %s\n", dbHost)
	}

	// Get integer value
	maxConnections, err := client.GetInt(ctx, "database.max_connections")
	if err != nil {
		log.Printf("Error getting max_connections: %v", err)
	} else {
		fmt.Printf("Max Connections: %d\n", maxConnections)
	}

	// Get boolean value
	debugMode, err := client.GetBool(ctx, "app.debug_mode")
	if err != nil {
		log.Printf("Error getting debug_mode: %v", err)
	} else {
		fmt.Printf("Debug Mode: %v\n", debugMode)
	}

	// Get with default value
	timeout := client.GetIntWithDefault(ctx, "request.timeout", 30)
	fmt.Printf("Request Timeout: %d seconds\n", timeout)

	// Get JSON config
	dbConfig, err := client.GetJSON(ctx, "database.config")
	if err != nil {
		log.Printf("Error getting database.config: %v", err)
	} else {
		fmt.Printf("Database Config: %+v\n", dbConfig)
	}

	// Example 2: Set config values
	fmt.Println("\n=== Writing Configs ===")

	// Set string
	err = client.Set(ctx, "app.version", "v2.0.0", "admin@example.com")
	if err != nil {
		log.Printf("Error setting app.version: %v", err)
	}

	// Set integer
	err = client.Set(ctx, "database.max_connections", 150, "admin@example.com")
	if err != nil {
		log.Printf("Error setting max_connections: %v", err)
	}

	// Set boolean
	err = client.Set(ctx, "app.debug_mode", false, "admin@example.com")
	if err != nil {
		log.Printf("Error setting debug_mode: %v", err)
	}

	// Set JSON object
	dbSettings := map[string]interface{}{
		"host":     "postgres-prod.example.com",
		"port":     5432,
		"ssl_mode": "require",
	}
	err = client.Set(ctx, "database.connection", dbSettings, "admin@example.com")
	if err != nil {
		log.Printf("Error setting database.connection: %v", err)
	}

	// Example 3: List all configs
	fmt.Println("\n=== Listing All Configs ===")
	allConfigs, err := client.ListAll(ctx)
	if err != nil {
		log.Printf("Error listing configs: %v", err)
	} else {
		for key, config := range allConfigs {
			fmt.Printf("%s = %v (type: %s)\n", key, config.Value, config.ValueType)
		}
	}

	// Example 4: Delete a config
	fmt.Println("\n=== Deleting Config ===")
	err = client.Delete(ctx, "old.deprecated.config", "admin@example.com")
	if err != nil {
		log.Printf("Error deleting config: %v", err)
	}

	// Example 5: Watch for config changes
	fmt.Println("\n=== Watching for changes ===")
	fmt.Println("Configs will auto-refresh every 30 seconds...")

	// Simulate watching
	time.Sleep(2 * time.Second)

	// Stop polling when done
	client.StopPolling()
}
