package config

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"time"
)

type Config struct {
	HTTP     HTTPConfig
	Database DatabaseConfig
	Redis    RedisConfig
}

type HTTPConfig struct {
	Addr              string
	RequestTimeout    time.Duration
	ReadHeaderTimeout time.Duration
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
}

type DatabaseConfig struct {
	URL             string
	MaxConns        int32
	MinConns        int32
	MaxConnLifetime time.Duration
	MaxConnIdleTime time.Duration
	ExecTimeout     time.Duration
}

type RedisConfig struct {
	Addr         string
	Password     string
	DB           int
	PoolSize     int
	MinIdleConns int
	ExecTimeout  time.Duration
}

func Load() (Config, error) {
	port := envString("PORT", "8080")
	requestTimeout, err := envDuration("HTTP_REQUEST_TIMEOUT", 5*time.Second)
	if err != nil {
		return Config{}, err
	}
	readHeaderTimeout, err := envDuration("HTTP_READ_HEADER_TIMEOUT", 2*time.Second)
	if err != nil {
		return Config{}, err
	}
	readTimeout, err := envDuration("HTTP_READ_TIMEOUT", 5*time.Second)
	if err != nil {
		return Config{}, err
	}
	writeTimeout, err := envDuration("HTTP_WRITE_TIMEOUT", 10*time.Second)
	if err != nil {
		return Config{}, err
	}
	idleTimeout, err := envDuration("HTTP_IDLE_TIMEOUT", 60*time.Second)
	if err != nil {
		return Config{}, err
	}
	maxConns, err := envInt("POSTGRES_POOL_MAX_CONNS", 8)
	if err != nil {
		return Config{}, err
	}
	minConns, err := envInt("POSTGRES_POOL_MIN_CONNS", 1)
	if err != nil {
		return Config{}, err
	}
	maxConnLifetime, err := envDuration("POSTGRES_POOL_MAX_CONN_LIFETIME", 30*time.Minute)
	if err != nil {
		return Config{}, err
	}
	maxConnIdleTime, err := envDuration("POSTGRES_POOL_MAX_CONN_IDLE_TIME", 5*time.Minute)
	if err != nil {
		return Config{}, err
	}
	execTimeout, err := envDuration("POSTGRES_EXEC_TIMEOUT", 3*time.Second)
	if err != nil {
		return Config{}, err
	}
	redisDB, err := envInt("REDIS_DB", 0)
	if err != nil {
		return Config{}, err
	}
	redisPoolSize, err := envInt("REDIS_POOL_SIZE", 8)
	if err != nil {
		return Config{}, err
	}
	redisMinIdleConns, err := envInt("REDIS_MIN_IDLE_CONNS", 1)
	if err != nil {
		return Config{}, err
	}
	redisExecTimeout, err := envDuration("REDIS_EXEC_TIMEOUT", 3*time.Second)
	if err != nil {
		return Config{}, err
	}

	cfg := Config{
		HTTP: HTTPConfig{
			Addr:              net.JoinHostPort("", port),
			RequestTimeout:    requestTimeout,
			ReadHeaderTimeout: readHeaderTimeout,
			ReadTimeout:       readTimeout,
			WriteTimeout:      writeTimeout,
			IdleTimeout:       idleTimeout,
		},
		Database: DatabaseConfig{
			URL:             envString("DATABASE_URL", ""),
			MaxConns:        int32(maxConns),
			MinConns:        int32(minConns),
			MaxConnLifetime: maxConnLifetime,
			MaxConnIdleTime: maxConnIdleTime,
			ExecTimeout:     execTimeout,
		},
		Redis: RedisConfig{
			Addr:         envString("REDIS_ADDR", "localhost:6379"),
			Password:     envString("REDIS_PASSWORD", ""),
			DB:           redisDB,
			PoolSize:     redisPoolSize,
			MinIdleConns: redisMinIdleConns,
			ExecTimeout:  redisExecTimeout,
		},
	}

	if cfg.Database.URL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.Database.MaxConns < 1 {
		return Config{}, fmt.Errorf("POSTGRES_POOL_MAX_CONNS must be at least 1")
	}
	if cfg.Database.MinConns < 0 || cfg.Database.MinConns > cfg.Database.MaxConns {
		return Config{}, fmt.Errorf("POSTGRES_POOL_MIN_CONNS must be between 0 and POSTGRES_POOL_MAX_CONNS")
	}
	if cfg.Redis.PoolSize < 1 {
		return Config{}, fmt.Errorf("REDIS_POOL_SIZE must be at least 1")
	}
	if cfg.Redis.MinIdleConns < 0 || cfg.Redis.MinIdleConns > cfg.Redis.PoolSize {
		return Config{}, fmt.Errorf("REDIS_MIN_IDLE_CONNS must be between 0 and REDIS_POOL_SIZE")
	}

	return cfg, nil
}

func envString(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func envInt(key string, fallback int) (int, error) {
	value := os.Getenv(key)
	if value == "" {
		return fallback, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer: %w", key, err)
	}
	return parsed, nil
}

func envDuration(key string, fallback time.Duration) (time.Duration, error) {
	value := os.Getenv(key)
	if value == "" {
		return fallback, nil
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("%s must be a duration: %w", key, err)
	}
	return parsed, nil
}
