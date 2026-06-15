package config

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"auto-pr-review/internal/llm"
)

type Config struct {
	HTTP            HTTPConfig
	Database        DatabaseConfig
	Redis           RedisConfig
	Kafka           KafkaConfig
	KafkaConsumer   KafkaConsumerConfig
	JobOrchestrator JobOrchestratorConfig
	Jobs            JobsConfig
	LLM             llm.Config
	GitHub          GitHubConfig
	PRReview        PRReviewConfig
}

type PRReviewerRuntimeConfig struct {
	LLM      llm.Config
	GitHub   GitHubConfig
	PRReview PRReviewConfig
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

type KafkaConfig struct {
	Brokers        []string
	Topic          string
	DLQTopic       string
	ClientID       string
	MaxRetries     int
	RetryBackoff   time.Duration
	MaxInFlight    int
	PublishTimeout time.Duration
}

type KafkaConsumerConfig struct {
	GroupID        string
	MaxRetries     int
	RetryBackoff   time.Duration
	AttemptTimeout time.Duration
	MaxInFlight    int
}

type JobOrchestratorConfig struct {
	Interval time.Duration
}

type JobsConfig struct {
	PeriodOptions []string
}

type GitHubConfig struct {
	TokenPath string
}

type PRReviewConfig struct {
	OutputDir string
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
	kafkaBrokers, err := envList("KAFKA_BROKERS", "kafka:9092")
	if err != nil {
		return Config{}, err
	}
	kafkaMaxRetries, err := envInt("KAFKA_MAX_RETRIES", 5)
	if err != nil {
		return Config{}, err
	}
	kafkaRetryBackoff, err := envDuration("KAFKA_RETRY_BACKOFF", 100*time.Millisecond)
	if err != nil {
		return Config{}, err
	}
	kafkaMaxInFlight, err := envInt("KAFKA_MAX_IN_FLIGHT", 10)
	if err != nil {
		return Config{}, err
	}
	kafkaPublishTimeout, err := envDuration("KAFKA_PUBLISH_TIMEOUT", 3*time.Second)
	if err != nil {
		return Config{}, err
	}
	kafkaConsumerGroupID := envString("KAFKA_CONSUMER_GROUP_ID", "auto_pr_review")
	kafkaConsumerMaxRetries, err := envInt("KAFKA_CONSUMER_MAX_RETRIES", 5)
	if err != nil {
		return Config{}, err
	}
	kafkaConsumerRetryBackoff, err := envDuration("KAFKA_CONSUMER_RETRY_BACKOFF", 100*time.Millisecond)
	if err != nil {
		return Config{}, err
	}
	kafkaConsumerAttemptTimeout, err := envDuration("KAFKA_CONSUMER_ATTEMPT_TIMEOUT", 5*time.Second)
	if err != nil {
		return Config{}, err
	}
	kafkaConsumerMaxInFlight, err := envInt("KAFKA_CONSUMER_MAX_IN_FLIGHT", 15)
	if err != nil {
		return Config{}, err
	}
	jobOrchestratorInterval, err := envDuration("JOB_ORCHESTRATOR_INTERVAL", time.Second)
	if err != nil {
		return Config{}, err
	}
	jobPeriodOptions, err := envDurationList("JOB_PERIOD_OPTIONS", "1s,10m,5h")
	if err != nil {
		return Config{}, err
	}
	prReviewerCfg, err := LoadPRReviewer()
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
		Kafka: KafkaConfig{
			Brokers:        kafkaBrokers,
			Topic:          envString("KAFKA_TOPIC", "pr_review"),
			DLQTopic:       envString("KAFKA_DLQ_TOPIC", "pr_review_dlq"),
			ClientID:       envString("KAFKA_CLIENT_ID", "auto-pr-review-job-orchestrator"),
			MaxRetries:     kafkaMaxRetries,
			RetryBackoff:   kafkaRetryBackoff,
			MaxInFlight:    kafkaMaxInFlight,
			PublishTimeout: kafkaPublishTimeout,
		},
		KafkaConsumer: KafkaConsumerConfig{
			GroupID:        kafkaConsumerGroupID,
			MaxRetries:     kafkaConsumerMaxRetries,
			RetryBackoff:   kafkaConsumerRetryBackoff,
			AttemptTimeout: kafkaConsumerAttemptTimeout,
			MaxInFlight:    kafkaConsumerMaxInFlight,
		},
		JobOrchestrator: JobOrchestratorConfig{
			Interval: jobOrchestratorInterval,
		},
		Jobs: JobsConfig{
			PeriodOptions: jobPeriodOptions,
		},
		LLM: prReviewerCfg.LLM,
		GitHub: GitHubConfig{
			TokenPath: envString("GITHUB_TOKEN_PATH", ""),
		},
		PRReview: prReviewerCfg.PRReview,
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
	if len(cfg.Kafka.Brokers) == 0 {
		return Config{}, fmt.Errorf("KAFKA_BROKERS must contain at least one broker")
	}
	if cfg.Kafka.MaxRetries < 0 {
		return Config{}, fmt.Errorf("KAFKA_MAX_RETRIES must be non-negative")
	}
	if cfg.Kafka.RetryBackoff < 0 {
		return Config{}, fmt.Errorf("KAFKA_RETRY_BACKOFF must be non-negative")
	}
	if cfg.Kafka.MaxInFlight < 1 {
		return Config{}, fmt.Errorf("KAFKA_MAX_IN_FLIGHT must be at least 1")
	}
	if cfg.Kafka.PublishTimeout <= 0 {
		return Config{}, fmt.Errorf("KAFKA_PUBLISH_TIMEOUT must be positive")
	}
	if cfg.KafkaConsumer.GroupID == "" {
		return Config{}, fmt.Errorf("KAFKA_CONSUMER_GROUP_ID is required")
	}
	if cfg.KafkaConsumer.MaxRetries < 0 {
		return Config{}, fmt.Errorf("KAFKA_CONSUMER_MAX_RETRIES must be non-negative")
	}
	if cfg.KafkaConsumer.RetryBackoff < 0 {
		return Config{}, fmt.Errorf("KAFKA_CONSUMER_RETRY_BACKOFF must be non-negative")
	}
	if cfg.KafkaConsumer.AttemptTimeout <= 0 {
		return Config{}, fmt.Errorf("KAFKA_CONSUMER_ATTEMPT_TIMEOUT must be positive")
	}
	if cfg.KafkaConsumer.MaxInFlight < 1 {
		return Config{}, fmt.Errorf("KAFKA_CONSUMER_MAX_IN_FLIGHT must be at least 1")
	}
	if cfg.JobOrchestrator.Interval <= 0 {
		return Config{}, fmt.Errorf("JOB_ORCHESTRATOR_INTERVAL must be positive")
	}
	if len(cfg.Jobs.PeriodOptions) == 0 {
		return Config{}, fmt.Errorf("JOB_PERIOD_OPTIONS must contain at least one duration")
	}
	if cfg.LLM.Provider != llm.ProviderClaude && cfg.LLM.Provider != llm.ProviderChatGPT {
		return Config{}, fmt.Errorf("LLM_PROVIDER must be one of: %s, %s", llm.ProviderClaude, llm.ProviderChatGPT)
	}
	if cfg.LLM.RequestTimeout <= 0 {
		return Config{}, fmt.Errorf("LLM_REQUEST_TIMEOUT must be positive")
	}
	if cfg.LLM.MaxRetries < 0 {
		return Config{}, fmt.Errorf("LLM_MAX_RETRIES must be non-negative")
	}
	if cfg.LLM.RetryBackoff < 0 {
		return Config{}, fmt.Errorf("LLM_RETRY_BACKOFF must be non-negative")
	}
	if cfg.LLM.MaxIdleConns < 1 {
		return Config{}, fmt.Errorf("LLM_MAX_IDLE_CONNS must be at least 1")
	}
	if cfg.LLM.MaxIdleConnsPerHost < 1 {
		return Config{}, fmt.Errorf("LLM_MAX_IDLE_CONNS_PER_HOST must be at least 1")
	}
	if cfg.LLM.MaxConnsPerHost < 1 {
		return Config{}, fmt.Errorf("LLM_MAX_CONNS_PER_HOST must be at least 1")
	}
	if strings.TrimSpace(cfg.PRReview.OutputDir) == "" {
		return Config{}, fmt.Errorf("PR_REVIEW_OUTPUT_DIR is required")
	}

	return cfg, nil
}

func LoadPRReviewer() (PRReviewerRuntimeConfig, error) {
	llmCfg, err := loadLLM()
	if err != nil {
		return PRReviewerRuntimeConfig{}, err
	}
	cfg := PRReviewerRuntimeConfig{
		LLM: llmCfg,
		GitHub: GitHubConfig{
			TokenPath: envString("GITHUB_TOKEN_PATH", ""),
		},
		PRReview: PRReviewConfig{
			OutputDir: envString("PR_REVIEW_OUTPUT_DIR", "reports"),
		},
	}
	if err := validateLLM(cfg.LLM); err != nil {
		return PRReviewerRuntimeConfig{}, err
	}
	if strings.TrimSpace(cfg.PRReview.OutputDir) == "" {
		return PRReviewerRuntimeConfig{}, fmt.Errorf("PR_REVIEW_OUTPUT_DIR is required")
	}
	return cfg, nil
}

func loadLLM() (llm.Config, error) {
	provider := strings.ToLower(strings.TrimSpace(envString("LLM_PROVIDER", llm.ProviderClaude)))
	var apiKey, model string
	switch provider {
	case llm.ProviderClaude:
		apiKey = strings.TrimSpace(envString("ANTHROPIC_API_KEY", ""))
		model = strings.TrimSpace(envString("ANTHROPIC_MODEL", ""))
	case llm.ProviderChatGPT:
		apiKey = strings.TrimSpace(envString("OPENAI_API_KEY", ""))
		model = strings.TrimSpace(envString("OPENAI_MODEL", ""))
	}

	requestTimeout, err := envDuration("LLM_REQUEST_TIMEOUT", time.Minute)
	if err != nil {
		return llm.Config{}, err
	}
	maxRetries, err := envInt("LLM_MAX_RETRIES", 2)
	if err != nil {
		return llm.Config{}, err
	}
	retryBackoff, err := envDuration("LLM_RETRY_BACKOFF", 500*time.Millisecond)
	if err != nil {
		return llm.Config{}, err
	}
	maxIdleConns, err := envInt("LLM_MAX_IDLE_CONNS", 10)
	if err != nil {
		return llm.Config{}, err
	}
	maxIdleConnsPerHost, err := envInt("LLM_MAX_IDLE_CONNS_PER_HOST", 10)
	if err != nil {
		return llm.Config{}, err
	}
	maxConnsPerHost, err := envInt("LLM_MAX_CONNS_PER_HOST", 10)
	if err != nil {
		return llm.Config{}, err
	}

	return llm.Config{
		Provider:            provider,
		APIKey:              apiKey,
		Model:               model,
		RequestTimeout:      requestTimeout,
		MaxRetries:          maxRetries,
		RetryBackoff:        retryBackoff,
		MaxIdleConns:        maxIdleConns,
		MaxIdleConnsPerHost: maxIdleConnsPerHost,
		MaxConnsPerHost:     maxConnsPerHost,
	}, nil
}

func validateLLM(cfg llm.Config) error {
	if cfg.Provider != llm.ProviderClaude && cfg.Provider != llm.ProviderChatGPT {
		return fmt.Errorf("LLM_PROVIDER must be one of: %s, %s", llm.ProviderClaude, llm.ProviderChatGPT)
	}
	if cfg.RequestTimeout <= 0 {
		return fmt.Errorf("LLM_REQUEST_TIMEOUT must be positive")
	}
	if cfg.MaxRetries < 0 {
		return fmt.Errorf("LLM_MAX_RETRIES must be non-negative")
	}
	if cfg.RetryBackoff < 0 {
		return fmt.Errorf("LLM_RETRY_BACKOFF must be non-negative")
	}
	if cfg.MaxIdleConns < 1 {
		return fmt.Errorf("LLM_MAX_IDLE_CONNS must be at least 1")
	}
	if cfg.MaxIdleConnsPerHost < 1 {
		return fmt.Errorf("LLM_MAX_IDLE_CONNS_PER_HOST must be at least 1")
	}
	if cfg.MaxConnsPerHost < 1 {
		return fmt.Errorf("LLM_MAX_CONNS_PER_HOST must be at least 1")
	}
	return nil
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

func envList(key, fallback string) ([]string, error) {
	value := envString(key, fallback)
	parts := strings.Split(value, ",")
	items := make([]string, 0, len(parts))
	for _, part := range parts {
		item := strings.TrimSpace(part)
		if item == "" {
			continue
		}
		items = append(items, item)
	}
	return items, nil
}

func envDurationList(key, fallback string) ([]string, error) {
	value := envString(key, fallback)
	parts := strings.Split(value, ",")
	durations := make([]string, 0, len(parts))
	seen := make(map[string]struct{}, len(parts))

	for _, part := range parts {
		duration := strings.TrimSpace(part)
		if duration == "" {
			continue
		}
		parsed, err := time.ParseDuration(duration)
		if err != nil || parsed < time.Second {
			return nil, fmt.Errorf("%s must contain positive durations of at least 1s: %q", key, duration)
		}
		if _, ok := seen[duration]; ok {
			continue
		}
		seen[duration] = struct{}{}
		durations = append(durations, duration)
	}

	return durations, nil
}
