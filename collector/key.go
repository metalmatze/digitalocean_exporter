package collector

import (
	"context"
	"fmt"
	"time"

	"github.com/digitalocean/godo"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

// KeyCollector collects metrics about ssh keys added to the account.
type KeyCollector struct {
	logger  log.Logger
	errors  *prometheus.CounterVec
	client  *godo.Client
	timeout time.Duration

	Key *prometheus.Desc
}

// NewKeyCollector returns a new KeyCollector.
func NewKeyCollector(logger log.Logger, errors *prometheus.CounterVec, client *godo.Client, timeout time.Duration) *KeyCollector {
	errors.WithLabelValues("key").Add(0)

	return &KeyCollector{
		logger:  logger,
		errors:  errors,
		client:  client,
		timeout: timeout,

		Key: prometheus.NewDesc(
			"digitalocean_key",
			"Information about keys in your digitalocean account",
			[]string{"id", "name", "fingerprint"},
			nil,
		),
	}
}

// Describe sends the super-set of all possible descriptors of metrics
// collected by this Collector.
func (c *KeyCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.Key
}

// Collect is called by the Prometheus registry when collecting metrics.
func (c *KeyCollector) Collect(ch chan<- prometheus.Metric) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()
	keys, _, err := c.client.Keys.List(ctx, nil)
	if err != nil {
		c.errors.WithLabelValues("key").Add(1)
		level.Warn(c.logger).Log(
			"msg", "can't list keys",
			"err", err,
		)
	}

	for _, key := range keys {
		ch <- prometheus.MustNewConstMetric(
			c.Key,
			prometheus.GaugeValue,
			1.0,
			fmt.Sprintf("%d", key.ID), key.Name, key.Fingerprint,
		)
	}
}
