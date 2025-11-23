package persistent

import (
	"go.opentelemetry.io/otel/attribute"
)

type PostgresConfig struct {
	Username string
	Name     string
	Password string
	Host     string
	Port     string

	Attributes []attribute.KeyValue
}
