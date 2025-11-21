package factory

import (
	"github.com/msyamsula/portofolio/other-works/design-pattern/bridge-factory/fpriority"
)

type Priority interface {
	Format(m string) string
}

const (
	PriorityDefault = iota
	PriorityInfo
	PriorityWarning
	PriorityCritical
)

type priorityConstructor func() Priority

var priorityRegistry = map[int]priorityConstructor{
	PriorityDefault:  func() Priority { return nil },
	PriorityInfo:     func() Priority { return fpriority.NewInfo() },
	PriorityWarning:  func() Priority { return fpriority.NewWarning() },
	PriorityCritical: func() Priority { return fpriority.NewCritical() },
}

func NewPriority(priorityType int) Priority {
	if c, ok := priorityRegistry[priorityType]; ok {
		return c()
	}

	return nil
}
