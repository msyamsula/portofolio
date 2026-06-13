package kafkaevent

import (
	_ "embed"
	"fmt"

	"github.com/linkedin/goavro/v2"
)

//go:embed schema/pr_review_event.avsc
var periodicJobSchema []byte

type PeriodicJobEnvelope struct {
	SchemaVersion int32
	GitHubLink    string
	Period        string
	NextTimestamp int64
	Error         string
}

func NewCodec() (*goavro.Codec, error) {
	codec, err := goavro.NewCodec(string(periodicJobSchema))
	if err != nil {
		return nil, fmt.Errorf("load avro codec: %w", err)
	}

	return codec, nil
}

func Marshal(codec *goavro.Codec, payload PeriodicJobEnvelope) ([]byte, error) {
	native := map[string]any{
		"schema_version": payload.SchemaVersion,
		"github_link":    payload.GitHubLink,
		"period":         payload.Period,
		"next_timestamp": payload.NextTimestamp,
		"error":          payload.Error,
	}

	binary, err := codec.BinaryFromNative(nil, native)
	if err != nil {
		return nil, fmt.Errorf("marshal kafka message: %w", err)
	}

	return binary, nil
}

func Unmarshal(codec *goavro.Codec, data []byte) (PeriodicJobEnvelope, error) {
	native, _, err := codec.NativeFromBinary(data)
	if err != nil {
		return PeriodicJobEnvelope{}, fmt.Errorf("unmarshal kafka message: %w", err)
	}
	nativeMap, ok := native.(map[string]any)
	if !ok {
		return PeriodicJobEnvelope{}, fmt.Errorf("unexpected kafka message shape")
	}

	getString := func(key string) (string, error) {
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

	getInt32 := func(key string) (int32, error) {
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

	getInt64 := func(key string) (int64, error) {
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

	schemaVersion, err := getInt32("schema_version")
	if err != nil {
		return PeriodicJobEnvelope{}, err
	}
	githubLink, err := getString("github_link")
	if err != nil {
		return PeriodicJobEnvelope{}, err
	}
	period, err := getString("period")
	if err != nil {
		return PeriodicJobEnvelope{}, err
	}
	nextTimestamp, err := getInt64("next_timestamp")
	if err != nil {
		return PeriodicJobEnvelope{}, err
	}
	messageError, err := getString("error")
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
