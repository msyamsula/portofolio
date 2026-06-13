package schemaregistry

import "testing"

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

func TestRegistryMarshalPeriodicJobEnvelopeV2WithIdempotencyKey(t *testing.T) {
	t.Parallel()

	registry, err := New()
	if err != nil {
		t.Fatalf("new registry: %v", err)
	}

	want := PeriodicJobEnvelope{
		SchemaVersion:  int32(PeriodicJobEnvelopeV2Version),
		GitHubLink:     "https://github.com/octocat/Hello-World",
		Period:         "10m",
		NextTimestamp:  123,
		Error:          "",
		IdempotencyKey: "dispatch-key-1",
	}

	data, err := registry.MarshalPeriodicJobEnvelope(want, PeriodicJobEnvelopeV2Version)
	if err != nil {
		t.Fatalf("marshal periodic job envelope: %v", err)
	}

	got, err := registry.UnmarshalPeriodicJobEnvelope(data, PeriodicJobEnvelopeV2Version)
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
