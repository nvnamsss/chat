package configs

import (
	"fmt"
	"os"
	"time"

	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v2"
)

// Config represents the application configuration
type Config struct {
	App      App      `yaml:"app"`
	Database Database `yaml:"database"`
	Kafka    Kafka    `yaml:"kafka"`
	LLM      LLM      `yaml:"llm"`
	JWT      JWT      `yaml:"jwt"`
}

// App holds application-specific configuration
type App struct {
	Name        string `yaml:"name" envconfig:"APP_NAME" default:"chat-service"`
	Host        string `yaml:"host" envconfig:"APP_HOST" default:"0.0.0.0"`
	Port        int    `yaml:"port" envconfig:"APP_PORT" default:"8080"`
	Environment string `yaml:"environment" envconfig:"APP_ENV" default:"development"`
	LogLevel    string `yaml:"logLevel" envconfig:"LOG_LEVEL" default:"info"`
}

// Database holds database configuration
type Database struct {
	Host     string `yaml:"host" envconfig:"DB_HOST" required:"true"`
	Port     int    `yaml:"port" envconfig:"DB_PORT" default:"5432"`
	User     string `yaml:"user" envconfig:"DB_USER" required:"true"`
	Password string `yaml:"password" envconfig:"DB_PASSWORD" required:"true"`
	Name     string `yaml:"name" envconfig:"DB_NAME" required:"true"`
	SSLMode  string `yaml:"sslMode" envconfig:"DB_SSL_MODE" default:"disable"`
}

type Postgres struct {
	Username          string `default:"root" envconfig:"POSTGRES_USER"`
	Password          string `default:"1" envconfig:"POSTGRES_PASSWORD"`
	Host              string `default:"127.0.0.1" envconfig:"POSTGRES_HOST"`
	Port              int    `default:"5432" envconfig:"POSTGRES_PORT"`
	Database          string `default:"actifs" envconfig:"POSTGRES_DB"`
	MaxOpenConnection int    `default:"10" envconfig:"POSTGRES_MAX_OPEN"`
	MaxIdleConnection int    `default:"10" envconfig:"POSTGRES_MAX_IDLE"`
	MaxLifeTime       int    `default:"24" envconfig:"POSTGRES_MAX_LIFETIME"`
	LogLevel          int    `default:"0" envconfig:"POSTGRES_LOG_LEVEL"`
}

func (postgres *Postgres) ConnectionString() string {
	v := fmt.Sprintf("user=%s dbname=%s host=%s port=%d sslmode=disable password=%s TimeZone=UTC",
		postgres.Username, postgres.Database, postgres.Host,
		postgres.Port, postgres.Password)
	return v
}

// Kafka holds Kafka configuration
type Kafka struct {
	Brokers       []string `yaml:"brokers" envconfig:"KAFKA_BROKERS" required:"true"`
	ConsumerGroup string   `yaml:"consumerGroup" envconfig:"KAFKA_CONSUMER_GROUP" default:"chat-service"`
	Topics        Topics   `yaml:"topics"`
}

// Topics holds Kafka topic configuration
type Topics struct {
	Chat    string `yaml:"chat" envconfig:"KAFKA_TOPIC_CHAT" default:"chat"`
	Message string `yaml:"message" envconfig:"KAFKA_TOPIC_MESSAGE" default:"message"`
}

// LLM holds LLM vendor service configuration
type LLM struct {
	BaseURL   string        `yaml:"baseUrl" envconfig:"LLM_BASE_URL" required:"true"`
	Timeout   time.Duration `yaml:"timeout" envconfig:"LLM_TIMEOUT" default:"30s"`
	Model     string        `yaml:"model" envconfig:"LLM_MODEL" default:"gpt-4"`
	MaxTokens int           `yaml:"maxTokens" envconfig:"LLM_MAX_TOKENS" default:"2048"`
	APIKey    string        `yaml:"apiKey" envconfig:"LLM_API_KEY" required:"true"`
}

// JWT holds JWT authentication configuration
type JWT struct {
	Secret    string        `yaml:"secret" envconfig:"JWT_SECRET" required:"true"`
	ExpiresIn time.Duration `yaml:"expiresIn" envconfig:"JWT_EXPIRES_IN" default:"24h"`
}

// AppConfig is the global application configuration
var AppConfig Config

// Load loads configuration from file and environment variables
func Load(configPath string) error {
	// Default configuration
	config := Config{}
	if err := envconfig.Process("", &config); err != nil {
		return nil
	}

	// Load from YAML file if provided
	if configPath != "" {
		file, err := os.Open(configPath)
		if err != nil {
			return fmt.Errorf("error opening config file: %w", err)
		}
		defer file.Close()

		decoder := yaml.NewDecoder(file)
		if err := decoder.Decode(&config); err != nil {
			return fmt.Errorf("error parsing config file: %w", err)
		}
	}

	// Override with environment variables
	if err := envconfig.Process("", &config); err != nil {
		return fmt.Errorf("error processing environment variables: %w", err)
	}

	// Set global configuration
	AppConfig = config
	return nil
}

// DSN returns the database connection string
func (db *Database) DSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		db.Host, db.Port, db.User, db.Password, db.Name, db.SSLMode)
}
