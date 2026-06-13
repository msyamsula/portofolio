package schemaregistry

import (
	"errors"
	"testing"
)

func TestRegistryMarshalPeriodicJobEnvelopeWithExplicitVersion(t *testing.T) {
	t.Parallel()

	registry, err := New()
	if err != nil {
		t.Fatalf("new registry: %v", err)
	}

	want := PeriodicJobEnvelope{
		SchemaVersion: int32(PeriodicJobEnvelopeV1Version),
		GitHubLink:    "https://github.com/octocat/Hello-World",
		Period:        "10m",
		NextTimestamp: 123,
		Error:         "",
	}

	data, err := registry.MarshalPeriodicJobEnvelope(want, PeriodicJobEnvelopeV1Version)
	if err != nil {
		t.Fatalf("marshal periodic job envelope: %v", err)
	}

	got, err := registry.UnmarshalPeriodicJobEnvelope(data, PeriodicJobEnvelopeV1Version)
	if err != nil {
		t.Fatalf("unmarshal periodic job envelope: %v", err)
	}
	if got != want {
		t.Fatalf("payload mismatch: got %+v want %+v", got, want)
	}
}

func TestRegistryUnmarshalPeriodicJobEnvelopeReadsLegacyRawAvro(t *testing.T) {
	t.Parallel()

	registry, err := New()
	if err != nil {
		t.Fatalf("new registry: %v", err)
	}
	schema, err := registry.schemaByTypeAndVersion(MessageTypePeriodicJobEnvelope, PeriodicJobEnvelopeV1Version)
	if err != nil {
		t.Fatalf("v1 schema is not registered: %v", err)
	}

	want := PeriodicJobEnvelope{
		SchemaVersion: int32(PeriodicJobEnvelopeV1Version),
		GitHubLink:    "https://github.com/octocat/Hello-World",
		Period:        "10m",
		NextTimestamp: 123,
		Error:         "",
	}
	legacyData, err := schema.Codec.BinaryFromNative(nil, map[string]any{
		"schema_version": want.SchemaVersion,
		"github_link":    want.GitHubLink,
		"period":         want.Period,
		"next_timestamp": want.NextTimestamp,
		"error":          want.Error,
	})
	if err != nil {
		t.Fatalf("marshal legacy payload: %v", err)
	}

	got, err := registry.UnmarshalPeriodicJobEnvelope(legacyData, PeriodicJobEnvelopeV1Version)
	if err != nil {
		t.Fatalf("unmarshal legacy payload: %v", err)
	}
	if got != want {
		t.Fatalf("payload mismatch: got %+v want %+v", got, want)
	}
}

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
