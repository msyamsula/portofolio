package schemaregistry

import (
	"errors"
	"testing"
)

func TestRegistryRejectsUnknownSchemaVersion(t *testing.T) {
	t.Parallel()

	registry, err := New()
	if err != nil {
		t.Fatalf("new registry: %v", err)
	}

	if _, err := registry.UnmarshalPeriodicJobEnvelope([]byte{1, 2, 3}, SchemaVersion(999)); !errors.Is(err, ErrInvalidAvro) {
		t.Fatalf("expected invalid avro error, got %v", err)
	}
}

func TestRegistryRejectsNilMessageTypeVersionMap(t *testing.T) {
	t.Parallel()

	registry, err := New()
	if err != nil {
		t.Fatalf("new registry: %v", err)
	}
	registry.schemasByType[MessageTypePeriodicJobEnvelope] = nil

	_, err = registry.MarshalPeriodicJobEnvelope(PeriodicJobEnvelope{}, PeriodicJobEnvelopeV1Version)
	if !errors.Is(err, ErrInvalidAvro) {
		t.Fatalf("expected invalid avro error, got %v", err)
	}
}

func TestRegistryRejectsNilSchemaDefinition(t *testing.T) {
	t.Parallel()

	registry, err := New()
	if err != nil {
		t.Fatalf("new registry: %v", err)
	}
	registry.schemasByType[MessageTypePeriodicJobEnvelope][PeriodicJobEnvelopeV1Version] = nil

	_, err = registry.MarshalPeriodicJobEnvelope(PeriodicJobEnvelope{}, PeriodicJobEnvelopeV1Version)
	if !errors.Is(err, ErrInvalidAvro) {
		t.Fatalf("expected invalid avro error, got %v", err)
	}
}

func TestRegistryRejectsNilCodec(t *testing.T) {
	t.Parallel()

	registry, err := New()
	if err != nil {
		t.Fatalf("new registry: %v", err)
	}
	registry.schemasByType[MessageTypePeriodicJobEnvelope][PeriodicJobEnvelopeV1Version].Codec = nil

	_, err = registry.UnmarshalPeriodicJobEnvelope([]byte{1, 2, 3}, PeriodicJobEnvelopeV1Version)
	if !errors.Is(err, ErrInvalidAvro) {
		t.Fatalf("expected invalid avro error, got %v", err)
	}
}

func TestRegistryRejectsMissingMessageType(t *testing.T) {
	t.Parallel()

	registry, err := New()
	if err != nil {
		t.Fatalf("new registry: %v", err)
	}
	delete(registry.schemasByType, MessageTypePeriodicJobEnvelope)

	_, err = registry.MarshalPeriodicJobEnvelope(PeriodicJobEnvelope{}, PeriodicJobEnvelopeV1Version)
	if !errors.Is(err, ErrInvalidAvro) {
		t.Fatalf("expected invalid avro error, got %v", err)
	}
}
