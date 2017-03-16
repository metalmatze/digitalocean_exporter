package collector

import (
	"context"
	"log"

	"github.com/digitalocean/godo"
	"github.com/prometheus/client_golang/prometheus"
)

type SnapshotCollector struct {
	client *godo.Client

	Size        *prometheus.Desc
	MinDiskSize *prometheus.Desc
}

func NewSnapshotCollector(client *godo.Client) *SnapshotCollector {
	labels := []string{"id", "name", "region", "type"}
	return &SnapshotCollector{
		client: client,

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

func (c *SnapshotCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.Size
}

func (c *SnapshotCollector) Collect(ch chan<- prometheus.Metric) {
	snapshots, _, err := c.client.Snapshots.List(context.TODO(), nil)
	if err != nil {
		log.Printf("Can't list snapshots: %v", err)
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
