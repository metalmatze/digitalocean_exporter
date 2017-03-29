package collector

import (
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

// ExporterCollector collects metrics, mostly runtime, about this exporter in general.
type ExporterCollector struct {
	logger    log.Logger
	version   string
	revision  string
	buildDate string
	goVersion string
	startTime time.Time

	StartTime *prometheus.Desc
	BuildInfo *prometheus.Desc
}

//logger, Version, Revision, BuildDate, GoVersion, StartTime

// NewExporterCollector returns a new ExporterCollector.
func NewExporterCollector(logger log.Logger, version string, revision string, buildDate string, goVersion string, startTime time.Time) *ExporterCollector {
	return &ExporterCollector{
		logger: logger,

		version:   version,
		revision:  revision,
		buildDate: buildDate,
		goVersion: goVersion,
		startTime: startTime,

		StartTime: prometheus.NewDesc(
			"digitalocean_start_time",
			"Unix timestamp of the start time",
			nil, nil,
		),
		BuildInfo: prometheus.NewDesc(
			"digitalocean_build_info",
			"A metric with a constant '1' value labeled by version, revision, and branch from which the node_exporter was built.",
			[]string{"verison", "revision", "builddate", "goversion"}, nil,
		),
	}
}

// Describe sends the super-set of all possible descriptors of metrics
// collected by this Collector.
func (c *ExporterCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.StartTime
}

// Collect is called by the Prometheus registry when collecting metrics.
func (c *ExporterCollector) Collect(ch chan<- prometheus.Metric) {
	level.Debug(c.logger).Log(
		"starttime", c.startTime.Unix(),
		"version", c.version,
		"revision", c.revision,
		"buildDate", c.buildDate,
		"goVersion", c.goVersion,
		"startTime", c.startTime,
	)

	ch <- prometheus.MustNewConstMetric(
		c.StartTime,
		prometheus.GaugeValue,
		float64(c.startTime.Unix()),
	)
	ch <- prometheus.MustNewConstMetric(
		c.BuildInfo,
		prometheus.GaugeValue,
		1.0,
		c.version, c.revision, c.buildDate, c.goVersion,
	)
}
