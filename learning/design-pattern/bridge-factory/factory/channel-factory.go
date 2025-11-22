package factory

import "github.com/msyamsula/portofolio/learning/design-pattern/bridge-factory/fchannel"

const (
	ChannelDefault = iota
	ChannelEmail
	ChannelFax
	ChannelPhone
)

type Channel interface {
	Send(m string) string
}

type channelConstructor func() Channel

var constructorRegistry = map[int]channelConstructor{
	ChannelDefault: func() Channel { return nil },
	ChannelEmail:   func() Channel { return fchannel.NewEmail() },
	ChannelFax:     func() Channel { return fchannel.NewFax() },
	ChannelPhone:   func() Channel { return fchannel.NewPhone() },
}

func NewChannel(channelType int) Channel {
	if c, ok := constructorRegistry[channelType]; ok {
		return c()
	}

	return nil
}
