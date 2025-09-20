package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Database  DatabaseConfig
	Redis     RedisConfig
	Server    ServerConfig
	External  ExternalConfig
	Scheduler SchedulerConfig
	Logger    LoggerConfig
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

type ServerConfig struct {
	Host string
	Port int
}

type ExternalConfig struct {
	MessageAPIURL string
	Timeout       time.Duration
}

type SchedulerConfig struct {
	Interval         time.Duration
	MessagesPerBatch int
}

type LoggerConfig struct {
	Level string
}

func Load() (*Config, error) {
	_ = godotenv.Load("config.env")

	cfg := &Config{
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvAsInt("DB_PORT", 5432),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", ""),
			DBName:   getEnv("DB_NAME", "insider_messaging"),
			SSLMode:  getEnv("DB_SSL_MODE", "disable"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnvAsInt("REDIS_PORT", 6379),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
		},
		Server: ServerConfig{
			Host: getEnv("SERVER_HOST", "0.0.0.0"),
			Port: getEnvAsInt("SERVER_PORT", 8080),
		},
		External: ExternalConfig{
			MessageAPIURL: getEnv("MESSAGE_API_URL", "https://webhook.site/your-webhook-id"),
			Timeout:       getEnvAsDuration("MESSAGE_API_TIMEOUT", 30*time.Second),
		},
		Scheduler: SchedulerConfig{
			Interval:         getEnvAsDuration("SCHEDULER_INTERVAL", 2*time.Minute),
			MessagesPerBatch: getEnvAsInt("MESSAGES_PER_BATCH", 2),
		},
		Logger: LoggerConfig{
			Level: getEnv("LOG_LEVEL", "info"),
		},
	}

	return cfg, nil
}

func (c *Config) GetDatabaseDSN() string {
	return "host=" + c.Database.Host +
		" port=" + strconv.Itoa(c.Database.Port) +
		" user=" + c.Database.User +
		" password=" + c.Database.Password +
		" dbname=" + c.Database.DBName +
		" sslmode=" + c.Database.SSLMode
}

func (c *Config) GetRedisAddr() string {
	return c.Redis.Host + ":" + strconv.Itoa(c.Redis.Port)
}

func (c *Config) GetServerAddr() string {
	return c.Server.Host + ":" + strconv.Itoa(c.Server.Port)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
