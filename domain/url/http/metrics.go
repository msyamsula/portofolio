package http

import "github.com/prometheus/client_golang/prometheus"

var (
	// prometheus metrics
	HashCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "hash_counter",
		Help: "number of shortener request",
	})

	RedirectCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "redirect_counter",
		Help: "number of redirect request",
	})
)
