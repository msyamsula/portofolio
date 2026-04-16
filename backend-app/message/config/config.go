package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds all runtime configuration loaded from environment variables.
// Every field has a sensible default so the service can run locally without
// any environment setup beyond a reachable Postgres and Kafka.
type Config struct {
	// DBDSN is the PostgreSQL connection string.
	DBDSN string

	// KafkaBrokers is a comma-free single broker address (host:port).
	// Multiple brokers can be supported by splitting on comma; kept simple here.
	KafkaBrokers string

	// KafkaTopic is the single topic used for all message events.
	KafkaTopic string

	// Port is the HTTP listen address, e.g. ":8080".
	Port string

	// MaxRetry is the maximum number of RECONCILE_STUCKED_MESSAGE events before
	// the reconciler escalates to RECONCILE_FAILED_MESSAGE attempts.
	MaxRetry int

	// MaxAttempt is the maximum number of RECONCILE_FAILED_MESSAGE events before
	// the reconciler writes a terminal FAILED event and stops processing the message.
	MaxAttempt int

	// OutboxBatch controls how many unpublished events the relay fetches per tick.
	OutboxBatch int

	// ReconcileInterval is how often the reconciler sweeps for stuck messages.
	ReconcileInterval time.Duration
}

// Load reads configuration from environment variables with fallback defaults.
func Load() Config {
	return Config{
		DBDSN:             env("DB_DSN", "postgres://msguser:msgpass@localhost:5432/messagedb"),
		KafkaBrokers:      env("KAFKA_BROKERS", "localhost:9092"),
		KafkaTopic:        env("KAFKA_TOPIC", "message.events"),
		Port:              env("PORT", ":8080"),
		MaxRetry:          envInt("MAX_RETRY", 3),
		MaxAttempt:        envInt("MAX_ATTEMPT", 3),
		OutboxBatch:       envInt("OUTBOX_BATCH", 100),
		ReconcileInterval: envDuration("RECONCILE_INTERVAL", time.Minute),
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func envDuration(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
}
