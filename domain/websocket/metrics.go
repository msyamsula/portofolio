package websocket

import "github.com/prometheus/client_golang/prometheus"

var (
	HubGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "hub_count",
		Help: "current active hub in the server",
	})

	UserGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "user_count",
		Help: "current active user in the server",
	})

	MessageCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "message_count",
		Help: "total message send over this server",
	})
)
