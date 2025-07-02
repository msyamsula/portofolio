package bridgefactory

import (
	"fmt"
	"testing"

	"github.com/msyamsula/portofolio/design_pattern/bridge_factory/factory"
)

func TestNotifications(t *testing.T) {
	n := NewNotification(factory.ChannelEmail, factory.PriorityCritical)
	fmt.Println(n.SendMessage("bridge and factory"))

	nn := NewNotification(factory.ChannelPhone, factory.PriorityWarning)
	fmt.Println(nn.SendMessage("bridge and Factory"))
}
