package schemaregistry

import (
	"errors"
	"fmt"

	"github.com/linkedin/goavro/v2"
)

var ErrInvalidAvro = errors.New("schema registry client invalid avro")

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

func New() (*Registry, error) {
	registry := &Registry{
		schemasByType: make(map[MessageType]map[SchemaVersion]*SchemaDefinition),
	}

	if err := registry.registerPeriodicJobEnvelopeSchemas(); err != nil {
		return nil, err
	}

	if err := registry.validate(); err != nil {
		return nil, err
	}

	return registry, nil
}

func (r *Registry) registerSchemas(messageType MessageType, schemas map[SchemaVersion]*SchemaDefinition) error {
	if r == nil {
		return invalidAvro("registry is nil")
	}
	if messageType == "" {
		return invalidAvro("message type is empty")
	}
	if schemas == nil {
		return invalidAvro("schema map for message type %s is nil", messageType)
	}
	if _, exists := r.schemasByType[messageType]; exists {
		return invalidAvro("message type %s is already registered", messageType)
	}

	r.schemasByType[messageType] = schemas
	return nil
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
