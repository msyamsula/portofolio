package adapter

import (
	"testing"

	"github.com/msyamsula/portofolio/other-works/design-pattern/adapter/local"
)

func TestAdapter(t *testing.T) {
	flog := local.NewAdapter(local.AdapterTypeFluentd)
	llog := local.NewAdapter(local.AdapterTypeLocal)
	slog := local.NewAdapter(local.AdapterTypeSyslog)

	message := "testing"
	flog.LogError(message)
	flog.LogInfo(message)

	llog.LogInfo(message)
	llog.LogError(message)

	slog.LogInfo(message)
	slog.LogError(message)
}
