package collector

import (
	"context"
	"fmt"
	"log"

	"github.com/digitalocean/godo"
	"github.com/prometheus/client_golang/prometheus"
)

// FloatingIPCollector collects metrics about all floating ips.
type FloatingIPCollector struct {
	client *godo.Client

	Active *prometheus.Desc
}

// NewFloatingIPCollector returns a new FloatingIPCollector.
func NewFloatingIPCollector(client *godo.Client) *FloatingIPCollector {
	labels := []string{"droplet_id", "droplet_name", "region", "ipv4"}

	return &FloatingIPCollector{
		client: client,

		Active: prometheus.NewDesc(
			"digitalocean_floating_ipv4_active",
			"If 1 the floating ip used by a droplet, 0 otherwise",
			labels, nil,
		),
	}
}

// Describe sends the super-set of all possible descriptors of metrics
// collected by this Collector.
func (c *FloatingIPCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.Active
}

// Collect is called by the Prometheus registry when collecting metrics.
func (c *FloatingIPCollector) Collect(ch chan<- prometheus.Metric) {
	floatingIPs, _, err := c.client.FloatingIPs.List(context.TODO(), nil)
	if err != nil {
		log.Printf("Can't list FloatingIPs: %v", err)
	}

	for _, ip := range floatingIPs {
		labels := []string{
			fmt.Sprintf("%d", ip.Droplet.ID),
			ip.Droplet.Name,
			ip.Region.Slug,
			ip.IP,
		}

		ch <- prometheus.MustNewConstMetric(
			c.Active,
			prometheus.GaugeValue,
			1.0,
			labels...,
		)
	}
}
