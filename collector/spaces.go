package collector

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/digitalocean/godo"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/prometheus/client_golang/prometheus"
)

// SpacesCollector collects metrics about all spaces buckets.
type SpacesCollector struct {
	logger          log.Logger
	errors          *prometheus.CounterVec
	client          *godo.Client
	timeout         time.Duration
	accessKeyID     string
	accessKeySecret string
	Bucket          *prometheus.Desc
	BucketCreated   *prometheus.Desc
}

// Templated since each region has a different endpoint
const spacesDomain = "%s.digitaloceanspaces.com"

// SpacesCollector returns a new SpacesCollector.
func NewSpacesCollector(logger log.Logger, errors *prometheus.CounterVec, client *godo.Client, accessKeyID string, accessKeySecret string, timeout time.Duration) *SpacesCollector {
	errors.WithLabelValues("spaces_bucket").Add(0)

	labels := []string{"region", "name"}
	return &SpacesCollector{
		logger:          logger,
		errors:          errors,
		client:          client,
		timeout:         timeout,
		accessKeyID:     accessKeyID,
		accessKeySecret: accessKeySecret,
		Bucket: prometheus.NewDesc(
			"digitalocean_spaces_bucket",
			"Spaces bucket and its details. Will always be 1 if exists",
			labels, nil,
		),
		BucketCreated: prometheus.NewDesc(
			"digitalocean_spaces_bucket_created",
			"Spaces bucket's creation date in unix epoch format",
			labels, nil,
		),
	}
}

// Describe sends the super-set of all possible descriptors of metrics
// collected by this Collector.
func (c *SpacesCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.Bucket
}

// Collect is called by the Prometheus registry when collecting metrics.
func (c *SpacesCollector) Collect(ch chan<- prometheus.Metric) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	regions, _, err := c.client.Regions.List(ctx, nil)
	if err != nil {
		c.errors.WithLabelValues("spaces_bucket").Add(1)
		level.Warn(c.logger).Log(
			"msg", "can't list regions",
			"err", err,
		)
		return
	}

	// The spaces API can be slow when checking each region 1 by 1, speed up by running them concurrently
	wg := sync.WaitGroup{}
	defer wg.Wait()

	for _, region := range regions {
		wg.Add(1)
		go func(region godo.Region) {
			defer wg.Done()
			spacesEndpoint := fmt.Sprintf(spacesDomain, region.Slug)

			spacesClient, err := minio.New(spacesEndpoint, &minio.Options{
				Creds:  credentials.NewStaticV4(c.accessKeyID, c.accessKeySecret, ""),
				Secure: true,
			})
			if err != nil {
				c.errors.WithLabelValues("spaces_bucket").Add(1)
				level.Warn(c.logger).Log(
					"msg", "can't create minio client",
					"err", err,
				)
				return
			}

			// Use a separate context than the godo client. The spaces API can be a bit slow
			buckets, err := spacesClient.ListBuckets(context.Background())
			if err != nil {
				// Not all regions may support spaces
				// Let's not log all of the known failures
				var dnsError *net.DNSError
				if errors.As(err, &dnsError) {
					return
				}
				c.errors.WithLabelValues("spaces_bucket").Add(1)
				level.Warn(c.logger).Log(
					"msg", "can't list spaces buckets",
					"err", err,
				)
				return
			}

			for _, bucket := range buckets {
				labels := []string{
					region.Slug,
					bucket.Name,
				}

				ch <- prometheus.MustNewConstMetric(
					c.Bucket,
					prometheus.GaugeValue,
					1.0,
					labels...,
				)

				ch <- prometheus.MustNewConstMetric(
					c.BucketCreated,
					prometheus.CounterValue,
					float64(bucket.CreationDate.Unix()),
					labels...,
				)
			}
		}(region)
	}
}
