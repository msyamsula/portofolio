package kafkaproducer

import (
	"context"
	_ "embed"
	"fmt"

	"auto-pr-review/internal/config"
	"auto-pr-review/internal/service"

	"github.com/IBM/sarama"
	"github.com/linkedin/goavro/v2"
)

//go:embed schema/pr_review_event.avsc
var periodicJobSchema []byte

type Producer struct {
	producer sarama.SyncProducer
	topic    string
	codec    *goavro.Codec
}

type periodicJobEnvelope struct {
	SchemaVersion int32  `json:"schema_version"`
	GitHubLink    string `json:"github_link"`
	Period        string `json:"period"`
	NextTimestamp int64  `json:"next_timestamp"`
	Error         string `json:"error"`
}

func NewProducer(cfg config.KafkaConfig) (*Producer, error) {
	codec, err := goavro.NewCodec(string(periodicJobSchema))
	if err != nil {
		return nil, fmt.Errorf("load avro codec: %w", err)
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
		producer: producer,
		topic:    cfg.Topic,
		codec:    codec,
	}, nil
}

func (p *Producer) Close() error {
	return p.producer.Close()
}

func (p *Producer) PublishPeriodicJob(ctx context.Context, job service.PeriodicJob) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}

	if err := p.publishOnce(p.topic, job.GitHubLink, periodicJobEnvelope{
		SchemaVersion: 1,
		GitHubLink:    job.GitHubLink,
		Period:        job.Period,
		NextTimestamp: job.NextTimestamp,
		Error:         "",
	}); err != nil {
		return false, err
	}

	return true, nil
}

func (p *Producer) publishOnce(topic, key string, payload periodicJobEnvelope) error {
	messageBody, err := p.marshal(payload)
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

func (p *Producer) marshal(payload periodicJobEnvelope) ([]byte, error) {
	native := map[string]any{
		"schema_version": payload.SchemaVersion,
		"github_link":    payload.GitHubLink,
		"period":         payload.Period,
		"next_timestamp": payload.NextTimestamp,
		"error":          payload.Error,
	}

	binary, err := p.codec.BinaryFromNative(nil, native)
	if err != nil {
		return nil, fmt.Errorf("marshal kafka message: %w", err)
	}

	return binary, nil
}
