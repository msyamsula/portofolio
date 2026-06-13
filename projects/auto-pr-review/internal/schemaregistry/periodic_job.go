package schemaregistry

import (
	_ "embed"
	"fmt"

	"github.com/linkedin/goavro/v2"
)

const (
	MessageTypePeriodicJobEnvelope MessageType = "periodic_job_envelope"

	PeriodicJobEnvelopeV1Version SchemaVersion = 1
	PeriodicJobEnvelopeV2Version SchemaVersion = 2
)

//go:embed schema/pr_review_event.avsc
var periodicJobSchemaV1 []byte

//go:embed schema/pr_review_event_v2.avsc
var periodicJobSchemaV2 []byte

type PeriodicJobEnvelope struct {
	SchemaVersion  int32
	GitHubLink     string
	Period         string
	NextTimestamp  int64
	Error          string
	IdempotencyKey string
}

func (r *Registry) registerPeriodicJobEnvelopeSchemas() error {
	schemas, err := newPeriodicJobEnvelopeSchemas()
	if err != nil {
		return err
	}

	return r.registerSchemas(MessageTypePeriodicJobEnvelope, schemas)
}

func newPeriodicJobEnvelopeSchemas() (map[SchemaVersion]*SchemaDefinition, error) {
	periodicJobV1Codec, err := goavro.NewCodec(string(periodicJobSchemaV1))
	if err != nil {
		return nil, fmt.Errorf("load periodic job v1 avro codec: %w", err)
	}
	periodicJobV2Codec, err := goavro.NewCodec(string(periodicJobSchemaV2))
	if err != nil {
		return nil, fmt.Errorf("load periodic job v2 avro codec: %w", err)
	}

	return map[SchemaVersion]*SchemaDefinition{
		PeriodicJobEnvelopeV1Version: {
			MessageType: MessageTypePeriodicJobEnvelope,
			Version:     PeriodicJobEnvelopeV1Version,
			Codec:       periodicJobV1Codec,
		},
		PeriodicJobEnvelopeV2Version: {
			MessageType: MessageTypePeriodicJobEnvelope,
			Version:     PeriodicJobEnvelopeV2Version,
			Codec:       periodicJobV2Codec,
		},
	}, nil
}

func (r *Registry) MarshalPeriodicJobEnvelope(payload PeriodicJobEnvelope, version SchemaVersion) ([]byte, error) {
	if payload.SchemaVersion == 0 {
		payload.SchemaVersion = int32(version)
	}

	schema, err := r.schemaByTypeAndVersion(MessageTypePeriodicJobEnvelope, version)
	if err != nil {
		return nil, err
	}

	native := map[string]any{
		"schema_version": payload.SchemaVersion,
		"github_link":    payload.GitHubLink,
		"period":         payload.Period,
		"next_timestamp": payload.NextTimestamp,
		"error":          payload.Error,
	}
	if version >= PeriodicJobEnvelopeV2Version {
		native["idempotency_key"] = payload.IdempotencyKey
	}

	body, err := schema.Codec.BinaryFromNative(nil, native)
	if err != nil {
		return nil, fmt.Errorf("marshal periodic job envelope: %w", err)
	}

	return body, nil
}

func (r *Registry) UnmarshalPeriodicJobEnvelope(data []byte, version SchemaVersion) (PeriodicJobEnvelope, error) {
	schema, err := r.schemaByTypeAndVersion(MessageTypePeriodicJobEnvelope, version)
	if err != nil {
		return PeriodicJobEnvelope{}, err
	}

	return r.unmarshalPeriodicJobEnvelope(schema, data)
}

func (r *Registry) unmarshalPeriodicJobEnvelope(schema *SchemaDefinition, data []byte) (PeriodicJobEnvelope, error) {
	if schema == nil || schema.Codec == nil {
		return PeriodicJobEnvelope{}, invalidAvro("periodic job schema is nil")
	}

	native, _, err := schema.Codec.NativeFromBinary(data)
	if err != nil {
		return PeriodicJobEnvelope{}, fmt.Errorf("unmarshal periodic job envelope: %w", err)
	}
	nativeMap, ok := native.(map[string]any)
	if !ok {
		return PeriodicJobEnvelope{}, fmt.Errorf("unexpected periodic job envelope shape")
	}

	schemaVersion, err := getInt32(nativeMap, "schema_version")
	if err != nil {
		return PeriodicJobEnvelope{}, err
	}
	githubLink, err := getString(nativeMap, "github_link")
	if err != nil {
		return PeriodicJobEnvelope{}, err
	}
	period, err := getString(nativeMap, "period")
	if err != nil {
		return PeriodicJobEnvelope{}, err
	}
	nextTimestamp, err := getInt64(nativeMap, "next_timestamp")
	if err != nil {
		return PeriodicJobEnvelope{}, err
	}
	messageError, err := getString(nativeMap, "error")
	if err != nil {
		return PeriodicJobEnvelope{}, err
	}
	idempotencyKey, err := getStringDefault(nativeMap, "idempotency_key", "")
	if err != nil {
		return PeriodicJobEnvelope{}, err
	}

	return PeriodicJobEnvelope{
		SchemaVersion:  schemaVersion,
		GitHubLink:     githubLink,
		Period:         period,
		NextTimestamp:  nextTimestamp,
		Error:          messageError,
		IdempotencyKey: idempotencyKey,
	}, nil
}
