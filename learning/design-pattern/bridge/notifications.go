package bridge

import (
	"github.com/msyamsula/portofolio/other-works/design-pattern/bridge/channel"
	"github.com/msyamsula/portofolio/other-works/design-pattern/bridge/priority"
)

type Notification struct {
	Priority priority.Priority
	Channel  channel.Channel
}

func (n *Notification) SendMessage(m string) string {
	return n.Channel.Send(n.Priority.Format(m))
}
