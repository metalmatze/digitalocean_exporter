package collector

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/digitalocean/godo"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

// DBCollector collects metrics about all databases.
type DBCollector struct {
	logger  log.Logger
	errors  *prometheus.CounterVec
	client  *godo.Client
	timeout time.Duration

	DB      *prometheus.Desc
	DBNodes *prometheus.Desc
}

// NewDBCollector returns a new DBCollector.
func NewDBCollector(logger log.Logger, errors *prometheus.CounterVec, client *godo.Client, timeout time.Duration) *DBCollector {
	errors.WithLabelValues("database").Add(0)
	labels := []string{
		"id",
		"name",
		"maintenance_window_day",
		"maintenance_window_hour",
		"maintenance_window_pending",
		"region",
		"size",
		"engine",
		"version",
	}
	return &DBCollector{
		logger:  logger,
		errors:  errors,
		client:  client,
		timeout: timeout,

		DB: prometheus.NewDesc(
			"digitalocean_database_status",
			"If 1 the database is online, 0 otherwise",
			labels, nil,
		),
		DBNodes: prometheus.NewDesc(
			"digitalocean_database_nodes",
			"Number of nodes in a database cluster",
			labels, nil,
		),
	}
}

// Describe sends the super-set of all possible descriptors of metrics
// collected by this Collector.
func (c *DBCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.DB
	ch <- c.DBNodes
}

// Collect is called by the Prometheus registry when collecting metrics.
func (c *DBCollector) Collect(ch chan<- prometheus.Metric) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	// create a list to hold our dbs
	dbs := []godo.Database{}

	// create options. initially, these will be blank
	opt := &godo.ListOptions{}

	for {
		dbPage, resp, err := c.client.Databases.List(ctx, opt)
		if err != nil {
			c.errors.WithLabelValues("database").Add(1)
			level.Warn(c.logger).Log(
				"msg", "can't list databases",
				"err", err,
			)
		}

		// append the current page's dbs to our list
		for _, d := range dbPage {
			dbs = append(dbs, d)
		}

		// if we are at the last page, break out the for loop
		if resp.Links == nil || resp.Links.IsLastPage() {
			break
		}

		page, err := resp.Links.CurrentPage()
		if err != nil {
			c.errors.WithLabelValues("database").Add(1)
			level.Warn(c.logger).Log(
				"msg", "can't read current page",
				"err", err,
			)
		}

		opt.Page = page + 1
	}

	for _, db := range dbs {
		labels := []string{
			db.ID,
			db.Name,
			db.MaintenanceWindow.Day,
			db.MaintenanceWindow.Hour,
			strconv.FormatBool(db.MaintenanceWindow.Pending),
			db.RegionSlug,
			db.SizeSlug,
			db.EngineSlug,
			db.VersionSlug,
		}
		var dbStatus float64
		// API gives back lowercase string already, this is for assurance
		if strings.ToLower(db.Status) == "online" {
			dbStatus = 1.0
		}
		ch <- prometheus.MustNewConstMetric(
			c.DB,
			prometheus.GaugeValue,
			dbStatus,
			labels...,
		)
		ch <- prometheus.MustNewConstMetric(
			c.DBNodes,
			prometheus.GaugeValue,
			float64(db.NumNodes),
			labels...,
		)
	}
}
