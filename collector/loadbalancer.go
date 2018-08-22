package collector

import (
	"context"
	"time"

	"github.com/digitalocean/godo"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

// LoadBalancerCollector collects metrics about LoadBalancers of that account.
type LoadBalancerCollector struct {
	logger  log.Logger
	errors  *prometheus.CounterVec
	client  *godo.Client
	timeout time.Duration

	Droplets *prometheus.Desc
	Status   *prometheus.Desc
}

// NewLoadBalancerCollector returns a new LoadBalancerCollector.
func NewLoadBalancerCollector(logger log.Logger, errors *prometheus.CounterVec, client *godo.Client, timeout time.Duration) *LoadBalancerCollector {
	errors.WithLabelValues("loadbalancer").Add(0)

	return &LoadBalancerCollector{
		logger:  logger,
		errors:  errors,
		client:  client,
		timeout: timeout,

		Droplets: prometheus.NewDesc(
			"digitalocean_loadbalancer_droplets",
			"The number of droplets this load balancer is proxying to",
			[]string{"id", "name", "ip"},
			nil,
		),
		Status: prometheus.NewDesc(
			"digitalocean_loadbalancer_status",
			"The status of the load balancer, 1 if active",
			[]string{"id", "name", "ip"},
			nil,
		),
	}
}

// Describe sends the super-set of all possible descriptors of metrics
// collected by this Collector.
func (c *LoadBalancerCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.Droplets
	ch <- c.Status
}

// Collect is called by the Prometheus registry when collecting metrics.
func (c *LoadBalancerCollector) Collect(ch chan<- prometheus.Metric) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	lbs, _, err := c.client.LoadBalancers.List(ctx, nil)
	if err != nil {
		c.errors.WithLabelValues("loadbalancer").Add(1)
		level.Warn(c.logger).Log(
			"msg", "can't list loadbalancers",
			"err", err,
		)
	}

	for _, lb := range lbs {
		status := 0.0
		if lb.Status == "active" {
			status = 1
		}

		ch <- prometheus.MustNewConstMetric(
			c.Status,
			prometheus.GaugeValue,
			status,
			lb.ID, lb.Name, lb.IP,
		)

		ch <- prometheus.MustNewConstMetric(
			c.Droplets,
			prometheus.GaugeValue,
			float64(len(lb.DropletIDs)),
			lb.ID, lb.Name, lb.IP,
		)
	}
}
