package schemaregistry

import (
	_ "embed"
	"errors"
	"fmt"

	"github.com/linkedin/goavro/v2"
)

var ErrInvalidAvro = errors.New("schema registry client invalid avro")

const (
	MessageTypePeriodicJobEnvelope MessageType = "periodic_job_envelope"

	PeriodicJobEnvelopeV1Version SchemaVersion = 1
)

//go:embed schema/pr_review_event.avsc
var periodicJobSchemaV1 []byte

type MessageType string

type SchemaVersion int32

type SchemaDefinition struct {
	MessageType MessageType
	Version     SchemaVersion
	Codec       *goavro.Codec
}

type Registry struct {
	schemasByType map[MessageType]map[SchemaVersion]*SchemaDefinition
}

type PeriodicJobEnvelope struct {
	SchemaVersion int32
	GitHubLink    string
	Period        string
	NextTimestamp int64
	Error         string
}

func New() (*Registry, error) {
	periodicJobV1Codec, err := goavro.NewCodec(string(periodicJobSchemaV1))
	if err != nil {
		return nil, fmt.Errorf("load periodic job v1 avro codec: %w", err)
	}

	periodicJobV1 := &SchemaDefinition{
		MessageType: MessageTypePeriodicJobEnvelope,
		Version:     PeriodicJobEnvelopeV1Version,
		Codec:       periodicJobV1Codec,
	}

	registry := &Registry{
		schemasByType: map[MessageType]map[SchemaVersion]*SchemaDefinition{
			MessageTypePeriodicJobEnvelope: {
				PeriodicJobEnvelopeV1Version: periodicJobV1,
			},
		},
	}

	if err := registry.validate(); err != nil {
		return nil, err
	}

	return registry, nil
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

	return PeriodicJobEnvelope{
		SchemaVersion: schemaVersion,
		GitHubLink:    githubLink,
		Period:        period,
		NextTimestamp: nextTimestamp,
		Error:         messageError,
	}, nil
}

func (r *Registry) validate() error {
	if r == nil {
		return invalidAvro("registry is nil")
	}
	for messageType, schemasByVersion := range r.schemasByType {
		if messageType == "" {
			return invalidAvro("message type is empty")
		}
		if schemasByVersion == nil {
			return invalidAvro("schema map for message type %s is nil", messageType)
		}
		for version, schema := range schemasByVersion {
			if schema == nil {
				return invalidAvro("schema for message type %s version %d is nil", messageType, version)
			}
			if schema.Codec == nil {
				return invalidAvro("codec for message type %s version %d is nil", messageType, version)
			}
			if schema.MessageType != messageType {
				return invalidAvro("message type mismatch for version %d: got %s want %s", version, schema.MessageType, messageType)
			}
			if schema.Version != version {
				return invalidAvro("schema version mismatch for message type %s: got %d want %d", messageType, schema.Version, version)
			}
		}
	}

	return nil
}

func (r *Registry) schemaByTypeAndVersion(messageType MessageType, version SchemaVersion) (*SchemaDefinition, error) {
	if r == nil {
		return nil, invalidAvro("registry is nil")
	}
	schemasByVersion, ok := r.schemasByType[messageType]
	if !ok {
		return nil, invalidAvro("message type %s is not registered", messageType)
	}
	if schemasByVersion == nil {
		return nil, invalidAvro("schema map for message type %s is nil", messageType)
	}
	schema, ok := schemasByVersion[version]
	if !ok {
		return nil, invalidAvro("schema version %d for message type %s is not registered", version, messageType)
	}
	if schema == nil {
		return nil, invalidAvro("schema for message type %s version %d is nil", messageType, version)
	}
	if schema.Codec == nil {
		return nil, invalidAvro("codec for message type %s version %d is nil", messageType, version)
	}

	return schema, nil
}

func invalidAvro(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrInvalidAvro, fmt.Sprintf(format, args...))
}

func getString(nativeMap map[string]any, key string) (string, error) {
	value, ok := nativeMap[key]
	if !ok {
		return "", fmt.Errorf("missing kafka field %s", key)
	}
	str, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("kafka field %s is not a string", key)
	}
	return str, nil
}

func getInt32(nativeMap map[string]any, key string) (int32, error) {
	value, ok := nativeMap[key]
	if !ok {
		return 0, fmt.Errorf("missing kafka field %s", key)
	}
	switch typed := value.(type) {
	case int32:
		return typed, nil
	case int:
		return int32(typed), nil
	default:
		return 0, fmt.Errorf("kafka field %s is not an int32", key)
	}
}

func getInt64(nativeMap map[string]any, key string) (int64, error) {
	value, ok := nativeMap[key]
	if !ok {
		return 0, fmt.Errorf("missing kafka field %s", key)
	}
	switch typed := value.(type) {
	case int64:
		return typed, nil
	case int32:
		return int64(typed), nil
	case int:
		return int64(typed), nil
	default:
		return 0, fmt.Errorf("kafka field %s is not an int64", key)
	}
}
