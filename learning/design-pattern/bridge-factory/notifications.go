package bridgefactory

import (
	"github.com/msyamsula/portofolio/other-works/design-pattern/bridge-factory/factory"
)

type Notification struct {
	Channel  factory.Channel
	Priority factory.Priority
}

func NewNotification(channelType, priorityType int) Notification {
	return Notification{
		Channel:  factory.NewChannel(channelType),
		Priority: factory.NewPriority(priorityType),
	}
}

func (n Notification) SendMessage(m string) string {
	return n.Channel.Send(n.Priority.Format(m))
}
