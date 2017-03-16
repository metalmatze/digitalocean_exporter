package collector

import (
	"context"
	"log"

	"github.com/digitalocean/godo"
	"github.com/prometheus/client_golang/prometheus"
)

type VolumeCollector struct {
	client *godo.Client

	Size *prometheus.Desc
}

func NewVolumeCollector(client *godo.Client) *VolumeCollector {
	labels := []string{"id", "name", "region"}
	return &VolumeCollector{
		client: client,

		Size: prometheus.NewDesc(
			"digitalocean_volume_size_bytes",
			"Volume's size in bytes",
			labels, nil,
		),
	}
}

func (c *VolumeCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.Size
}

func (c *VolumeCollector) Collect(ch chan<- prometheus.Metric) {
	volumes, _, err := c.client.Storage.ListVolumes(context.TODO(), nil)
	if err != nil {
		log.Printf("Can't list volumes: %v", err)
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
