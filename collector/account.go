package collector

import (
	"context"
	"time"

	"github.com/digitalocean/godo"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

// AccountCollector collects metrics about the account.
type AccountCollector struct {
	logger  log.Logger
	errors  *prometheus.CounterVec
	client  *godo.Client
	timeout time.Duration

	DropletLimit    *prometheus.Desc
	FloatingIPLimit *prometheus.Desc
	EmailVerified   *prometheus.Desc
	Active          *prometheus.Desc
}

// NewAccountCollector returns a new AccountCollector.
func NewAccountCollector(logger log.Logger, errors *prometheus.CounterVec, client *godo.Client, timeout time.Duration) *AccountCollector {
	errors.WithLabelValues("account").Add(0)

	return &AccountCollector{
		logger:  logger,
		errors:  errors,
		client:  client,
		timeout: timeout,

		DropletLimit: prometheus.NewDesc(
			"digitalocean_account_droplet_limit",
			"The maximum number of droplet you can use",
			nil, nil,
		),
		FloatingIPLimit: prometheus.NewDesc(
			"digitalocean_account_floating_ip_limit",
			"The maximum number of floating ips you can use",
			nil, nil,
		),
		EmailVerified: prometheus.NewDesc(
			"digitalocean_account_verified",
			"1 if your email address was verified",
			nil, nil,
		),
		Active: prometheus.NewDesc(
			"digitalocean_account_active",
			"The status of your account",
			nil, nil,
		),
	}
}

// Describe sends the super-set of all possible descriptors of metrics
// collected by this Collector.
func (c *AccountCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.DropletLimit
	ch <- c.FloatingIPLimit
	ch <- c.EmailVerified
	ch <- c.Active
}

// Collect is called by the Prometheus registry when collecting metrics.
func (c *AccountCollector) Collect(ch chan<- prometheus.Metric) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()
	acc, _, err := c.client.Account.Get(ctx)
	if err != nil {
		c.errors.WithLabelValues("account").Add(1)
		level.Warn(c.logger).Log(
			"msg", "can't get account",
			"err", err,
		)
		return
	}

	ch <- prometheus.MustNewConstMetric(
		c.DropletLimit,
		prometheus.GaugeValue,
		float64(acc.DropletLimit),
	)
	ch <- prometheus.MustNewConstMetric(
		c.FloatingIPLimit,
		prometheus.GaugeValue,
		float64(acc.FloatingIPLimit),
	)

	var verified float64
	if acc.EmailVerified {
		verified = 1
	}
	ch <- prometheus.MustNewConstMetric(
		c.EmailVerified,
		prometheus.GaugeValue,
		verified,
	)

	var status float64
	if acc.Status == "active" {
		status = 1
	}
	ch <- prometheus.MustNewConstMetric(
		c.Active,
		prometheus.GaugeValue,
		status,
	)
}
