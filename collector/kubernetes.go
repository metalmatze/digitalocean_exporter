package collector

import (
	"context"
	"time"

	"github.com/digitalocean/godo"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

// KubernetesCollector collects metrics about Kubernetes clusters
type KubernetesCollector struct {
	logger  log.Logger
	errors  *prometheus.CounterVec
	client  *godo.Client
	timeout time.Duration

	Up        *prometheus.Desc
	NodePools *prometheus.Desc
	Nodes     *prometheus.Desc
}

// NewKubernetesCollector returns a new KubernetesCollector
func NewKubernetesCollector(logger log.Logger, errors *prometheus.CounterVec, client *godo.Client, timeout time.Duration) *KubernetesCollector {
	errors.WithLabelValues("kubernetes").Add(0)

	// Version refers to the upstream Kubernetes version as well as the DigitalOcean revision
	clusterLabels := []string{"id", "name", "region", "version"}
	nodeLabels := []string{"id", "name", "region"}
	return &KubernetesCollector{
		logger:  logger,
		errors:  errors,
		client:  client,
		timeout: timeout,

		Up: prometheus.NewDesc(
			"digitalocean_kubernetes_cluster_up",
			"If 1 the kubernetes cluster is up and running, 0 otherwise",
			clusterLabels, nil,
		),
		NodePools: prometheus.NewDesc(
			"digitalocean_kubernetes_nodepools_count",
			"Number of Kubernetes nodepools",
			clusterLabels, nil,
		),
		Nodes: prometheus.NewDesc(
			"digitalocean_kubernetes_nodes_count",
			"Number of Kubernetes nodes",
			nodeLabels, nil,
		),
	}
}

// Describe secnds the super-set of all possible descriptors of metrics collected by this Collector.
func (c *KubernetesCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.Up
	ch <- c.NodePools
	ch <- c.Nodes
}

// Collect is called by the Prometheus registry when collecting metrics
func (c *KubernetesCollector) Collect(ch chan<- prometheus.Metric) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()
	clusters, _, err := c.client.Kubernetes.List(ctx, nil)
	if err != nil {
		c.errors.WithLabelValues("kubernetes").Add(1)
		level.Warn(c.logger).Log(
			"msg", "can't list clusters",
			"err", err,
		)
	}

	for _, cluster := range clusters {
		labels := []string{
			cluster.ID,
			cluster.Name,
			cluster.RegionSlug,
			cluster.VersionSlug,
		}

		var active float64
		//TODO(dazwilkin) better reflect richer Kubernetes cluster states
		if cluster.Status.State == godo.KubernetesClusterStatusRunning {
			active = 1.0
		}
		ch <- prometheus.MustNewConstMetric(
			c.Up,
			prometheus.GaugeValue,
			active,
			labels...,
		)
		ch <- prometheus.MustNewConstMetric(
			c.NodePools,
			prometheus.GaugeValue,
			float64(len(cluster.NodePools)),
			labels...,
		)

		for _, nodepool := range cluster.NodePools {
			// Assume NodePools are constrained to the cluster's Region
			// If so, we can labels a cluster's NodePools by the cluster's region
			labels := []string{
				nodepool.ID,
				nodepool.Name,
				cluster.RegionSlug,
			}
			ch <- prometheus.MustNewConstMetric(
				c.Nodes,
				prometheus.GaugeValue,
				float64(nodepool.Count),
				labels...,
			)
		}
	}
}
