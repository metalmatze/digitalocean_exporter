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

// DomainCollector collects metrics about all images created by the user.
type DomainCollector struct {
	logger  log.Logger
	client  *godo.Client
	timeout time.Duration

	DomainRecordPort     *prometheus.Desc
	DomainRecordPriority *prometheus.Desc
	DomainRecordWeight   *prometheus.Desc
	DomainTTL            *prometheus.Desc
}

// NewDomainCollector returns a new DomainCollector.
func NewDomainCollector(logger log.Logger, client *godo.Client, timeout time.Duration) *DomainCollector {
	recordLabels := []string{"id", "name", "type", "data"}

	return &DomainCollector{
		logger:  logger,
		client:  client,
		timeout: timeout,

		DomainRecordPort: prometheus.NewDesc(
			"digitalocean_domain_record_port",
			"The port for SRV records",
			recordLabels, nil,
		),
		DomainRecordPriority: prometheus.NewDesc(
			"digitalocean_domain_record_priority",
			"The priority for SRV and MX records",
			recordLabels, nil,
		),
		DomainRecordWeight: prometheus.NewDesc(
			"digitalocean_domain_record_weight",
			"The weight for SRV records",
			recordLabels, nil,
		),
		DomainTTL: prometheus.NewDesc(
			"digitalocean_domain_ttl_seconds",
			"Seconds that clients can cache queried information before a refresh should be requested",
			[]string{"name"}, nil,
		),
	}
}

// Describe sends the super-set of all possible descriptors of metrics
// collected by this Collector.
func (c *DomainCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.DomainRecordPort
	ch <- c.DomainRecordPriority
	ch <- c.DomainRecordWeight
	ch <- c.DomainTTL
}

// Collect is called by the Prometheus registry when collecting metrics.
func (c *DomainCollector) Collect(ch chan<- prometheus.Metric) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	domains, _, err := c.client.Domains.List(ctx, nil)
	if err != nil {
		level.Warn(c.logger).Log(
			"msg", "can't list domains",
			"err", err,
		)
		return
	}

	for _, domain := range domains {
		ch <- prometheus.MustNewConstMetric(
			c.DomainTTL,
			prometheus.GaugeValue,
			float64(domain.TTL),
			domain.Name,
		)

		//ctx, cancel := context.WithTimeout(ctx, c.timeout)
		//cancel()
		ctx := context.TODO()
		records, _, _ := c.client.Domains.Records(ctx, domain.Name, nil)
		for _, record := range records {
			ch <- prometheus.MustNewConstMetric(
				c.DomainRecordPort,
				prometheus.GaugeValue,
				float64(record.Port),
				fmt.Sprintf("%d", record.ID), record.Name, record.Type, record.Data,
			)
			ch <- prometheus.MustNewConstMetric(
				c.DomainRecordPriority,
				prometheus.GaugeValue,
				float64(record.Priority),
				fmt.Sprintf("%d", record.ID), record.Name, record.Type, record.Data,
			)
			ch <- prometheus.MustNewConstMetric(
				c.DomainRecordWeight,
				prometheus.GaugeValue,
				float64(record.Weight),
				fmt.Sprintf("%d", record.ID), record.Name, record.Type, record.Data,
			)
		}
	}
}
