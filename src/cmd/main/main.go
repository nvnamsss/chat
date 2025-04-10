package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/nvnamsss/chat/src/adapters"
	"github.com/nvnamsss/chat/src/configs"
	"github.com/nvnamsss/chat/src/controllers"
	"github.com/nvnamsss/chat/src/dtos"
	"github.com/nvnamsss/chat/src/logger"
	"github.com/nvnamsss/chat/src/middlewares"
	"github.com/nvnamsss/chat/src/repositories"
	"github.com/nvnamsss/chat/src/services"
)

func main() {
	// Parse command-line flags
	configPath := flag.String("config", "", "path to config file")
	flag.Parse()

	// Load configuration
	if err := configs.Load(*configPath); err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}
	cfg := configs.AppConfig
	log.Printf("Loaded config: %+v", cfg)

	// Initialize logger
	logger.Init(cfg.App.LogLevel, cfg.App.Environment)
	defer logger.Sync()

	// Set Gin mode
	if cfg.App.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Connect to database
	dbAdapter, err := adapters.NewDBAdapter(cfg.Database)
	if err != nil {
		logger.Fatal("Failed to connect to database", logger.Field("error", err))
	}
	defer dbAdapter.Close()

	// Run database migrations
	if err := runMigrations(dbAdapter, "/home/namnv/git/chat/src/migrations"); err != nil {
		logger.Fatal("Failed to run migrations", logger.Field("error", err))
	}

	// Initialize Kafka producer
	kafkaProducer := setupKafka(cfg)

	// Initialize LLM adapter
	// llmAdapter := adapters.NewLLMAdapter(cfg.LLM)
	llmAdapter := adapters.NewNothingLLMAdapter()

	// Initialize repositories
	chatRepo := repositories.NewChatRepository(dbAdapter)
	messageRepo := repositories.NewMessageRepository(dbAdapter)

	// Initialize services
	chatService := services.NewChatService(chatRepo, kafkaProducer)
	messageService := services.NewMessageService(messageRepo, chatRepo, llmAdapter, kafkaProducer)

	// Initialize controllers
	chatController := controllers.NewChatController(chatService)
	messageController := controllers.NewMessageController(messageService, chatService)

	// Create router
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middlewares.Logger())
	router.Use(middlewares.RequestID())
	router.Use(middlewares.CORS())
	router.Use(middlewares.Auth(cfg.JWT.Secret))

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// API routes
	api := router.Group("/api/v1")
	{
		chatController.RegisterRoutes(api)
		messageController.RegisterRoutes(api)
	}

	// Start the server
	addr := fmt.Sprintf("%s:%d", cfg.App.Host, cfg.App.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	// Run the server in a goroutine
	go func() {
		logger.Info("Starting server",
			logger.Field("host", cfg.App.Host),
			logger.Field("port", cfg.App.Port),
			logger.Field("env", cfg.App.Environment))

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", logger.Field("error", err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Create a deadline to wait for current operations to complete
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", logger.Field("error", err))
	}

	logger.Info("Server exited")
}

// runMigrations runs database migrations
func runMigrations(dbAdapter adapters.DBAdapter, migrationsPath string) error {
	db := dbAdapter.GetDB().DB
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migration driver: %w", err)
	}

	// Add "file://" prefix to make it a proper URL with scheme
	m, err := migrate.NewWithDatabaseInstance(
		"file://"+migrationsPath,
		"postgres", driver)
	if err != nil {
		return fmt.Errorf("failed to create migration instance: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

// setupKafka initializes the Kafka producer
func setupKafka(cfg configs.Config) services.KafkaProducer {
	// In a real application, this would initialize a Kafka client
	// For simplicity, we'll use a mock implementation
	return &mockKafkaProducer{}
}

// mockKafkaProducer is a simple mock implementation of the KafkaProducer interface
type mockKafkaProducer struct{}

func (m *mockKafkaProducer) PublishChatEvent(ctx context.Context, message *dtos.KafkaMessage[dtos.ChatPayload]) error {
	logger.Context(ctx).Infow("Mock: Publishing chat event",
		"event", message.Event,
		"chatID", message.Payload.ChatID)
	return nil
}

func (m *mockKafkaProducer) PublishMessageEvent(ctx context.Context, message *dtos.KafkaMessage[dtos.MessagePayload]) error {
	logger.Context(ctx).Infow("Mock: Publishing message event",
		"event", message.Event,
		"messageID", message.Payload.MessageID,
		"chatID", message.Payload.ChatID)
	return nil
}
