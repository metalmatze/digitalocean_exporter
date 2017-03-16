package collector

import (
	"context"
	"log"

	"github.com/digitalocean/godo"
	"github.com/prometheus/client_golang/prometheus"
)

// AccountCollector collects metrics about the account.
type AccountCollector struct {
	client *godo.Client

	DropletLimit    *prometheus.Desc
	FloatingIPLimit *prometheus.Desc
	EmailVerified   *prometheus.Desc
	Status          *prometheus.Desc
}

// NewAccountCollector returns a new AccountCollector.
func NewAccountCollector(client *godo.Client) *AccountCollector {
	return &AccountCollector{
		client: client,

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
		Status: prometheus.NewDesc(
			"digitalocean_account_status",
			"The status of your account",
			[]string{"status"}, nil,
		),
	}
}

// Describe sends the super-set of all possible descriptors of metrics
// collected by this Collector.
func (c *AccountCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.DropletLimit
	ch <- c.FloatingIPLimit
	ch <- c.EmailVerified
	ch <- c.Status
}

// Collect is called by the Prometheus registry when collecting metrics.
func (c *AccountCollector) Collect(ch chan<- prometheus.Metric) {
	acc, _, err := c.client.Account.Get(context.TODO())
	if err != nil {
		log.Printf("Can't get account: %v", err)
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
		c.Status,
		prometheus.GaugeValue,
		status,
		acc.Status,
	)
}
