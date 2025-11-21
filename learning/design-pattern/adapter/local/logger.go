package local

import "github.com/msyamsula/portofolio/other-works/design-pattern/adapter/fixed"

type constructor func() Logger

const (
	AdapterTypeDefault = iota
	AdapterTypeFluentd
	AdapterTypeLocal
	AdapterTypeSyslog
)

var loggerRegistry = map[int]constructor{
	AdapterTypeDefault: func() Logger { return nil },
	AdapterTypeFluentd: func() Logger {
		return &FluendAdapter{
			fluend: &fixed.FluentD{},
		}
	},
	AdapterTypeLocal: func() Logger {
		return &LocalAdapter{
			logger: &fixed.LocalLog{},
		}
	},
	AdapterTypeSyslog: func() Logger {
		return &SysLogAdapter{
			logger: &fixed.SysLog{},
		}
	},
}

type Logger interface {
	LogInfo(message string)
	LogError(message string)
}

func NewAdapter(adapterType int) Logger {
	if c, ok := loggerRegistry[adapterType]; ok {
		return c()
	}

	return nil
}
