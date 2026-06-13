package kafkaproducer

import (
	"context"
	"fmt"

	"auto-pr-review/internal/config"
	"auto-pr-review/internal/kafkaevent"
	"auto-pr-review/internal/service"

	"github.com/IBM/sarama"
	"github.com/linkedin/goavro/v2"
)

type Producer struct {
	producer sarama.SyncProducer
	topic    string
	codec    *goavro.Codec
}

func NewProducer(cfg config.KafkaConfig) (*Producer, error) {
	codec, err := kafkaevent.NewCodec()
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

	if err := p.PublishEvent(ctx, p.topic, job.GitHubLink, kafkaevent.PeriodicJobEnvelope{
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

func (p *Producer) PublishEvent(ctx context.Context, topic, key string, payload kafkaevent.PeriodicJobEnvelope) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	return p.publishOnce(topic, key, payload)
}

func (p *Producer) publishOnce(topic, key string, payload kafkaevent.PeriodicJobEnvelope) error {
	messageBody, err := kafkaevent.Marshal(p.codec, payload)
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
