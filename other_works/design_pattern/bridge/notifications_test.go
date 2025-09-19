package bridge

import (
	"fmt"
	"testing"

	"github.com/msyamsula/portofolio/other_works/design_pattern/bridge/channel"
	"github.com/msyamsula/portofolio/other_works/design_pattern/bridge/priority"
)

func TestNotifications(t *testing.T) {
	notification := &Notification{
		Priority: priority.Info{},
		Channel:  channel.Phone{},
	}

	fmt.Println(notification.SendMessage("halo"))

	notification.Priority = priority.Critical{}
	notification.Channel = channel.Fax{}

	fmt.Println(notification.SendMessage("mantap"))

}
