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

// MonitoringCollector collects metrics about all droplets.
type MonitoringCollector struct {
	logger  log.Logger
	errors  *prometheus.CounterVec
	client  *godo.Client
	timeout time.Duration

	CPUMetrics             *prometheus.Desc
	MemoryTotalMetrics     *prometheus.Desc
	MemoryFreeMetrics      *prometheus.Desc
	MemoryAvailableMetrics *prometheus.Desc
	MemoryCachedMetrics    *prometheus.Desc
}

// NewMonitoringCollector returns a new DropletCollector.
func NewMonitoringCollector(logger log.Logger, errors *prometheus.CounterVec, client *godo.Client, timeout time.Duration) *MonitoringCollector {
	errors.WithLabelValues("droplet").Add(0)

	labels := []string{"id", "name", "region"}
	return &MonitoringCollector{
		logger:  logger,
		errors:  errors,
		client:  client,
		timeout: timeout,

		CPUMetrics: prometheus.NewDesc(
			"digitalocean_monitoring_cpu",
			"Droplet's CPU metrics in seconds",
			append(labels, "mode"), nil,
		),
		MemoryTotalMetrics: prometheus.NewDesc(
			"digitalocean_monitoring_memory_total",
			"Droplet's total memory metrics in bytes",
			labels, nil,
		),
		MemoryFreeMetrics: prometheus.NewDesc(
			"digitalocean_monitoring_memory_free",
			"Droplet's free memory metrics in bytes",
			labels, nil,
		),
		MemoryAvailableMetrics: prometheus.NewDesc(
			"digitalocean_monitoring_memory_available",
			"Droplet's available memory metrics in bytes",
			labels, nil,
		),
		MemoryCachedMetrics: prometheus.NewDesc(
			"digitalocean_monitoring_memory_cached",
			"Droplet's cached memory metrics in bytes",
			labels, nil,
		),
	}
}

// Describe sends the super-set of all possible descriptors of metrics
// collected by this Collector.
func (c *MonitoringCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.CPUMetrics
	ch <- c.MemoryTotalMetrics
	ch <- c.MemoryFreeMetrics
	ch <- c.MemoryAvailableMetrics
	ch <- c.MemoryCachedMetrics
}

// Collect is called by the Prometheus registry when collecting metrics.
func (c *MonitoringCollector) Collect(ch chan<- prometheus.Metric) {
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

		metricsReq := &godo.DropletMetricsRequest{
			HostID: fmt.Sprintf("%d", droplet.ID),
			Start:  time.Now().Add(-5 * time.Minute),
			End:    time.Now(),
		}

		resp, _, err := c.client.Monitoring.GetDropletCPU(ctx, metricsReq)
		if err != nil {
			c.errors.WithLabelValues("droplet").Add(1)
			level.Warn(c.logger).Log(
				"msg", "can't read current droplet CPU metrics",
				"err", err,
			)
		}

		for _, metric := range resp.Data.Result {
			lastValue := metric.Values[len(metric.Values)-1].Value
			mode := fmt.Sprintf("%s", metric.Metric["mode"])
			CPULabels := append(labels, mode)
			ch <- prometheus.MustNewConstMetric(
				c.CPUMetrics,
				prometheus.GaugeValue,
				float64(lastValue),
				CPULabels...,
			)
		}

		resp, _, err = c.client.Monitoring.GetDropletTotalMemory(ctx, metricsReq)
		if err != nil {
			c.errors.WithLabelValues("droplet").Add(1)
			level.Warn(c.logger).Log(
				"msg", "can't read current droplet total memory metrics",
				"err", err,
			)
		}

		for _, metric := range resp.Data.Result {
			lastValue := metric.Values[len(metric.Values)-1].Value
			ch <- prometheus.MustNewConstMetric(
				c.MemoryTotalMetrics,
				prometheus.GaugeValue,
				float64(lastValue),
				labels...,
			)
		}

		resp, _, err = c.client.Monitoring.GetDropletAvailableMemory(ctx, metricsReq)
		if err != nil {
			c.errors.WithLabelValues("droplet").Add(1)
			level.Warn(c.logger).Log(
				"msg", "can't read current droplet available memory metrics",
				"err", err,
			)
		}

		for _, metric := range resp.Data.Result {
			lastValue := metric.Values[len(metric.Values)-1].Value
			ch <- prometheus.MustNewConstMetric(
				c.MemoryAvailableMetrics,
				prometheus.GaugeValue,
				float64(lastValue),
				labels...,
			)
		}

		resp, _, err = c.client.Monitoring.GetDropletFreeMemory(ctx, metricsReq)
		if err != nil {
			c.errors.WithLabelValues("droplet").Add(1)
			level.Warn(c.logger).Log(
				"msg", "can't read current droplet free memory metrics",
				"err", err,
			)
		}

		for _, metric := range resp.Data.Result {
			lastValue := metric.Values[len(metric.Values)-1].Value
			ch <- prometheus.MustNewConstMetric(
				c.MemoryFreeMetrics,
				prometheus.GaugeValue,
				float64(lastValue),
				labels...,
			)
		}

		resp, _, err = c.client.Monitoring.GetDropletCachedMemory(ctx, metricsReq)
		if err != nil {
			c.errors.WithLabelValues("droplet").Add(1)
			level.Warn(c.logger).Log(
				"msg", "can't read current droplet cached memory metrics",
				"err", err,
			)
		}

		for _, metric := range resp.Data.Result {
			lastValue := metric.Values[len(metric.Values)-1].Value
			ch <- prometheus.MustNewConstMetric(
				c.MemoryCachedMetrics,
				prometheus.GaugeValue,
				float64(lastValue),
				labels...,
			)
		}
	}
}
