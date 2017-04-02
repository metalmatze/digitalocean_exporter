package collector

import (
	"context"
	"time"

	"github.com/digitalocean/godo"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

// SnapshotCollector collects metrics about all snapshots of droplets & volumes.
type SnapshotCollector struct {
	logger  log.Logger
	client  *godo.Client
	timeout time.Duration

	Size        *prometheus.Desc
	MinDiskSize *prometheus.Desc
}

// NewSnapshotCollector returns a new SnapshotCollector.
func NewSnapshotCollector(logger log.Logger, client *godo.Client, timeout time.Duration) *SnapshotCollector {
	labels := []string{"id", "name", "region", "type"}
	return &SnapshotCollector{
		logger:  logger,
		client:  client,
		timeout: timeout,

		Size: prometheus.NewDesc(
			"digitalocean_snapshot_size_bytes",
			"Snapshot's size in bytes",
			labels, nil,
		),
		MinDiskSize: prometheus.NewDesc(
			"digitalocean_snapshot_min_disk_size_bytes",
			"Minimum disk size for a droplet/volume to run this snapshot on in bytes",
			labels, nil,
		),
	}
}

// Describe sends the super-set of all possible descriptors of metrics
// collected by this Collector.
func (c *SnapshotCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.Size
}

// Collect is called by the Prometheus registry when collecting metrics.
func (c *SnapshotCollector) Collect(ch chan<- prometheus.Metric) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()
	snapshots, _, err := c.client.Snapshots.List(ctx, nil)
	if err != nil {
		level.Warn(c.logger).Log(
			"msg", "can't list snapshots",
			"err", err,
		)
		return
	}

	for _, snapshot := range snapshots {
		labels := []string{
			snapshot.ID,
			snapshot.Name,
			snapshot.Regions[0],
			snapshot.ResourceType,
		}

		ch <- prometheus.MustNewConstMetric(
			c.MinDiskSize,
			prometheus.GaugeValue,
			float64(snapshot.MinDiskSize*1024*1024*1024),
			labels...,
		)

		if snapshot.SizeGigaBytes > 0 {
			ch <- prometheus.MustNewConstMetric(
				c.Size,
				prometheus.GaugeValue,
				float64(snapshot.SizeGigaBytes*1024*1024*1024),
				labels...,
			)
		}
	}
}
