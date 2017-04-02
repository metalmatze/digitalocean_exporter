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

// FloatingIPCollector collects metrics about all floating ips.
type FloatingIPCollector struct {
	logger  log.Logger
	client  *godo.Client
	timeout time.Duration

	Active *prometheus.Desc
}

// NewFloatingIPCollector returns a new FloatingIPCollector.
func NewFloatingIPCollector(logger log.Logger, client *godo.Client, timeout time.Duration) *FloatingIPCollector {
	labels := []string{"droplet_id", "droplet_name", "region", "ipv4"}

	return &FloatingIPCollector{
		logger:  logger,
		client:  client,
		timeout: timeout,

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
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()
	floatingIPs, _, err := c.client.FloatingIPs.List(ctx, nil)
	if err != nil {
		level.Warn(c.logger).Log(
			"msg", "can't list floating ips",
			"err", err,
		)
		return
	}

	for _, ip := range floatingIPs {
		var active float64
		var dropletID, dropletName string
		if ip.Droplet != nil {
			active = 1
			dropletID = fmt.Sprintf("%d", ip.Droplet.ID)
			dropletName = ip.Droplet.Name
		}

		labels := []string{
			dropletID,
			dropletName,
			ip.Region.Slug,
			ip.IP,
		}

		ch <- prometheus.MustNewConstMetric(
			c.Active,
			prometheus.GaugeValue,
			active,
			labels...,
		)
	}
}
