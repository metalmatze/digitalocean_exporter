package collector

import (
	"context"
	"fmt"
	"log"

	"github.com/digitalocean/godo"
	"github.com/prometheus/client_golang/prometheus"
)

type KeyCollector struct {
	client *godo.Client

	Key *prometheus.Desc
}

func NewKeyCollector(client *godo.Client) *KeyCollector {
	return &KeyCollector{
		client: client,

		Key: prometheus.NewDesc(
			"digitalocean_key",
			"Information about keys in your digitalocean account",
			[]string{"id", "name", "fingerprint"},
			nil,
		),
	}
}

func (c *KeyCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.Key
}

func (c *KeyCollector) Collect(ch chan<- prometheus.Metric) {
	keys, _, err := c.client.Keys.List(context.TODO(), nil)
	if err != nil {
		log.Printf("Can't list keys: %v", err)
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
