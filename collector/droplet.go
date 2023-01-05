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

// DropletCollector collects metrics about all droplets.
type DropletCollector struct {
	logger  log.Logger
	errors  *prometheus.CounterVec
	client  *godo.Client
	timeout time.Duration

	Up           *prometheus.Desc
	CPUs         *prometheus.Desc
	Memory       *prometheus.Desc
	Disk         *prometheus.Desc
	PriceHourly  *prometheus.Desc
	PriceMonthly *prometheus.Desc
}

// NewDropletCollector returns a new DropletCollector.
func NewDropletCollector(logger log.Logger, errors *prometheus.CounterVec, client *godo.Client, timeout time.Duration) *DropletCollector {
	errors.WithLabelValues("droplet").Add(0)

	labels := []string{"id", "name", "region"}
	return &DropletCollector{
		logger:  logger,
		errors:  errors,
		client:  client,
		timeout: timeout,

		Up: prometheus.NewDesc(
			"digitalocean_droplet_up",
			"If 1 the droplet is up and running, 0 otherwise",
			labels, nil,
		),
		CPUs: prometheus.NewDesc(
			"digitalocean_droplet_cpus",
			"Droplet's number of CPUs",
			labels, nil,
		),
		Memory: prometheus.NewDesc(
			"digitalocean_droplet_memory_bytes",
			"Droplet's memory in bytes",
			labels, nil,
		),
		Disk: prometheus.NewDesc(
			"digitalocean_droplet_disk_bytes",
			"Droplet's disk in bytes",
			labels, nil,
		),
		PriceHourly: prometheus.NewDesc(
			"digitalocean_droplet_price_hourly",
			"Price of the Droplet billed hourly in dollars",
			labels, nil,
		),
		PriceMonthly: prometheus.NewDesc(
			"digitalocean_droplet_price_monthly",
			"Price of the Droplet billed monthly in dollars",
			labels, nil,
		),
	}
}

// Describe sends the super-set of all possible descriptors of metrics
// collected by this Collector.
func (c *DropletCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.Up
	ch <- c.CPUs
	ch <- c.Memory
	ch <- c.Disk
	ch <- c.PriceHourly
	ch <- c.PriceMonthly
}

// Collect is called by the Prometheus registry when collecting metrics.
func (c *DropletCollector) Collect(ch chan<- prometheus.Metric) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	// create a list to hold our droplets
	droplets := []godo.Droplet{}

	// create options. initially, these will be blank
	opt := &godo.ListOptions{}

	for {
		dropletsPage, resp, err := c.client.Droplets.List(ctx, opt)
		if err != nil {
			c.errors.WithLabelValues("droplet").Add(1)
			level.Warn(c.logger).Log(
				"msg", "can't list droplets",
				"err", err,
			)
			return
		}

		// append the current page's droplets to our list
		for _, d := range dropletsPage {
			droplets = append(droplets, d)
		}

		// if we are at the last page, break out the for loop
		if resp.Links == nil || resp.Links.IsLastPage() {
			break
		}

		page, err := resp.Links.CurrentPage()
		if err != nil {
			c.errors.WithLabelValues("droplet").Add(1)
			level.Warn(c.logger).Log(
				"msg", "can't read current page",
				"err", err,
			)
		}

		opt.Page = page + 1
	}

	for _, droplet := range droplets {
		labels := []string{
			fmt.Sprintf("%d", droplet.ID),
			droplet.Name,
			droplet.Region.Slug,
		}

		var active float64
		if droplet.Status == "active" {
			active = 1.0
		}
		ch <- prometheus.MustNewConstMetric(
			c.Up,
			prometheus.GaugeValue,
			active,
			labels...,
		)
		ch <- prometheus.MustNewConstMetric(
			c.CPUs,
			prometheus.GaugeValue,
			float64(droplet.Vcpus),
			labels...,
		)
		ch <- prometheus.MustNewConstMetric(
			c.Memory,
			prometheus.GaugeValue,
			float64(droplet.Memory*1024*1024),
			labels...,
		)
		ch <- prometheus.MustNewConstMetric(
			c.Disk,
			prometheus.GaugeValue,
			float64(droplet.Disk*1000*1000*1000),
			labels...,
		)
		ch <- prometheus.MustNewConstMetric(
			c.PriceHourly,
			prometheus.GaugeValue,
			float64(droplet.Size.PriceHourly),
			labels...,
		)
		ch <- prometheus.MustNewConstMetric(
			c.PriceMonthly,
			prometheus.GaugeValue,
			float64(droplet.Size.PriceMonthly),
			labels...,
		)
	}
}
