package main

import (
	"fmt"
	"log"
	"net/http"

	_ "github.com/lib/pq"

	"github.com/sachin/config-manager/internal/config"
	repository "github.com/sachin/config-manager/internal/repository/postgres"
	"github.com/sachin/config-manager/internal/service"
	transporthttp "github.com/sachin/config-manager/internal/transport/http"
)

func main() {
	log.Println("Starting config-service...")

	// Load configuration
	cfg := config.Load()
	log.Printf("Config loaded - Server: %s:%s, DB: %s:%s/%s",
		cfg.Server.Host, cfg.Server.Port,
		cfg.Database.Host, cfg.Database.Port, cfg.Database.DBName)

	// Connect to database
	log.Println("Connecting to database...")
	db, err := repository.Connect(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer repository.Close()

	log.Println("Connected to database successfully")

	// Initialize repositories
	serviceRepo := repository.NewServiceRepository(db)
	envRepo := repository.NewEnvironmentRepository(db)
	configRepo := repository.NewConfigRepository(db)
	auditRepo := repository.NewAuditRepository(db)

	// Initialize services
	serviceService := service.NewServiceService(serviceRepo)
	envService := service.NewEnvironmentService(envRepo, serviceRepo)
	configService := service.NewConfigService(configRepo, envRepo, auditRepo)
	rollbackService := service.NewRollbackService(configRepo, envRepo, auditRepo)
	auditService := service.NewAuditService(auditRepo)

	// Initialize handlers
	handlers := &transporthttp.Handlers{
		Service:     transporthttp.NewServiceHandler(serviceService),
		Environment: transporthttp.NewEnvironmentHandler(envService),
		Config:      transporthttp.NewConfigHandler(configService, rollbackService),
		Audit:       transporthttp.NewAuditHandler(auditService),
	}

	// Setup routes
	router := transporthttp.SetupRoutes(handlers)

	// Start server
	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	log.Printf("Server listening on %s", addr)

	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
