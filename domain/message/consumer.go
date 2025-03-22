package message

import "github.com/nsqio/go-nsq"

type Consumer struct {
	topic, channel string
	lookupds       []string
	nsqConfig      *nsq.Config

	handler     nsq.Handler
	nsqConsumer *nsq.Consumer
}

type Config struct {
	Name      string
	Lookupds  []string
	NsqConfig *nsq.Config
}

func getTopicAndChannel(name string) (string, string) {
	switch name {
	case ConfigSaveMessage:
		return TopicSaveMessage, ChannelSaveMessage
	}

	return "", ""
}

func New(cfg Config, handler nsq.Handler) (*Consumer, error) {
	topic, channel := getTopicAndChannel(cfg.Name)
	c := &Consumer{
		topic:     topic,
		channel:   channel,
		lookupds:  cfg.Lookupds,
		nsqConfig: cfg.NsqConfig,
		handler:   handler,
	}

	return c, nil
}

func (c *Consumer) Start() error {
	nsqConsumer, err := nsq.NewConsumer(c.topic, c.channel, c.nsqConfig)
	if err != nil {
		return err
	}

	nsqConsumer.AddConcurrentHandlers(c.handler, 3)
	c.nsqConsumer = nsqConsumer

	err = nsqConsumer.ConnectToNSQLookupds(c.lookupds)
	if err != nil {
		return err
	}

	return nil
}

func (c *Consumer) Stop() {
	c.nsqConsumer.Stop()
}
