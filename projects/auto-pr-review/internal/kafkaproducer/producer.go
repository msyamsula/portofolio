package kafkaproducer

import (
	"context"
	"fmt"

	"auto-pr-review/internal/config"
	"auto-pr-review/internal/schemaregistry"
	"auto-pr-review/internal/service"

	"github.com/IBM/sarama"
)

type Producer struct {
	producer       sarama.SyncProducer
	topic          string
	schemaRegistry *schemaregistry.Registry
}

func NewProducer(cfg config.KafkaConfig) (*Producer, error) {
	registry, err := schemaregistry.New()
	if err != nil {
		return nil, fmt.Errorf("load schema registry: %w", err)
	}

	saramaConfig := sarama.NewConfig()
	saramaConfig.Version = sarama.V2_8_0_0
	saramaConfig.ClientID = cfg.ClientID
	saramaConfig.Net.MaxOpenRequests = cfg.MaxInFlight
	saramaConfig.Producer.Retry.Max = cfg.MaxRetries
	saramaConfig.Producer.Retry.Backoff = cfg.RetryBackoff
	saramaConfig.Producer.Timeout = cfg.PublishTimeout
	saramaConfig.Producer.RequiredAcks = sarama.WaitForAll
	saramaConfig.Producer.Partitioner = sarama.NewHashPartitioner
	saramaConfig.Producer.Return.Successes = true

	client, err := sarama.NewClient(cfg.Brokers, saramaConfig)
	if err != nil {
		return nil, fmt.Errorf("create kafka client: %w", err)
	}

	producer, err := sarama.NewSyncProducerFromClient(client)
	if err != nil {
		if closeErr := client.Close(); closeErr != nil {
			return nil, fmt.Errorf("create kafka producer: %w; close client: %w", err, closeErr)
		}
		return nil, fmt.Errorf("create kafka producer: %w", err)
	}

	return &Producer{
		producer:       producer,
		topic:          cfg.Topic,
		schemaRegistry: registry,
	}, nil
}

func (p *Producer) Close() error {
	return p.producer.Close()
}

func (p *Producer) PublishPeriodicJob(ctx context.Context, job service.PeriodicJob) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}

	version := schemaregistry.PeriodicJobEnvelopeV2Version
	if err := p.PublishEvent(ctx, p.topic, job.GitHubLink, schemaregistry.PeriodicJobEnvelope{
		SchemaVersion:  int32(version),
		GitHubLink:     job.GitHubLink,
		Period:         job.Period,
		NextTimestamp:  job.NextTimestamp,
		Error:          "",
		IdempotencyKey: job.IdempotencyKey,
	}, version); err != nil {
		return false, err
	}

	return true, nil
}

func (p *Producer) PublishEvent(ctx context.Context, topic, key string, payload schemaregistry.PeriodicJobEnvelope, version schemaregistry.SchemaVersion) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	return p.publishOnce(topic, key, payload, version)
}

func (p *Producer) publishOnce(topic, key string, payload schemaregistry.PeriodicJobEnvelope, version schemaregistry.SchemaVersion) error {
	messageBody, err := p.schemaRegistry.MarshalPeriodicJobEnvelope(payload, version)
	if err != nil {
		return err
	}

	_, _, err = p.producer.SendMessage(&sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.ByteEncoder(messageBody),
	})
	if err != nil {
		return fmt.Errorf("send kafka message: %w", err)
	}

	return nil
}
