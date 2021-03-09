package collector

import (
	"context"
	"time"

	"github.com/digitalocean/godo"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

// AppCollector collects metrics about all apps.
type AppCollector struct {
	logger  log.Logger
	errors  *prometheus.CounterVec
	client  *godo.Client
	timeout time.Duration

	App *prometheus.Desc
}

// NewAppCollector returns a new AppCollector.
func NewAppCollector(logger log.Logger, errors *prometheus.CounterVec, client *godo.Client, timeout time.Duration) *AppCollector {
	errors.WithLabelValues("app").Add(0)

	labels := []string{"id", "name", "tier", "region", "phase"}
	return &AppCollector{
		logger:  logger,
		errors:  errors,
		client:  client,
		timeout: timeout,

		App: prometheus.NewDesc(
			"digitalocean_app",
			"Information about an app deployed on the app platform",
			labels, nil,
		),
	}
}

// Describe sends the super-set of all possible descriptors of metrics
// collected by this Collector.
func (c *AppCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.App
}

// Collect is called by the Prometheus registry when collecting metrics.
func (c *AppCollector) Collect(ch chan<- prometheus.Metric) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	// create a list to hold our apps
	apps := []godo.App{}

	// create options. initially, these will be blank
	opt := &godo.ListOptions{}

	for {
		appsPage, resp, err := c.client.Apps.List(ctx, opt)
		if err != nil {
			c.errors.WithLabelValues("app").Add(1)
			level.Warn(c.logger).Log(
				"msg", "can't list apps",
				"err", err,
			)
		}

		// append the current page's apps to our list
		for _, a := range appsPage {
			apps = append(apps, *a)
		}

		// if we are at the last page, break out the for loop
		if resp.Links == nil || resp.Links.IsLastPage() {
			break
		}

		page, err := resp.Links.CurrentPage()
		if err != nil {
			c.errors.WithLabelValues("app").Add(1)
			level.Warn(c.logger).Log(
				"msg", "can't read current page",
				"err", err,
			)
		}

		opt.Page = page + 1
	}

	for _, app := range apps {
		// Need to check for an active deployment otherwise exporter will panic
		// Panic will occur if app was just created and hasn't had initial deployment yet
		if app.ActiveDeployment == nil {
			continue
		}
		phase := string(app.ActiveDeployment.Phase)
		// If this struct is populated, deployment is occurring
		if app.InProgressDeployment != nil {
			phase = string(app.InProgressDeployment.Phase)
		}
		labels := []string{
			app.ID,
			app.Spec.Name,
			app.TierSlug,
			app.Region.Slug,
			phase,
		}
		ch <- prometheus.MustNewConstMetric(
			c.App,
			prometheus.GaugeValue,
			1.0,
			labels...,
		)
	}
}
