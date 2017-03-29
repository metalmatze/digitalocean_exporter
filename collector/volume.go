package collector

import (
	"context"

	"github.com/digitalocean/godo"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

// VolumeCollector collects metrics about all volumes.
type VolumeCollector struct {
	logger log.Logger
	client *godo.Client

	Size *prometheus.Desc
}

// NewVolumeCollector returns a new VolumeCollector.
func NewVolumeCollector(logger log.Logger, client *godo.Client) *VolumeCollector {
	labels := []string{"id", "name", "region"}
	return &VolumeCollector{
		logger: logger,
		client: client,

		Size: prometheus.NewDesc(
			"digitalocean_volume_size_bytes",
			"Volume's size in bytes",
			labels, nil,
		),
	}
}

// Describe sends the super-set of all possible descriptors of metrics
// collected by this Collector.
func (c *VolumeCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.Size
}

// Collect is called by the Prometheus registry when collecting metrics.
func (c *VolumeCollector) Collect(ch chan<- prometheus.Metric) {
	volumes, _, err := c.client.Storage.ListVolumes(context.TODO(), nil)
	if err != nil {
		level.Warn(c.logger).Log(
			"msg", "can't list volumes",
			"err", err,
		)
	}

	for _, vol := range volumes {
		labels := []string{
			vol.ID,
			vol.Name,
			vol.Region.Slug,
		}

		ch <- prometheus.MustNewConstMetric(
			c.Size,
			prometheus.GaugeValue,
			float64(vol.SizeGigaBytes*1024*1024*1024),
			labels...,
		)
	}
}
