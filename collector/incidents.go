package collector

import (
	"errors"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/digitalocean/godo"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/tidwall/gjson"
)

const doStatusAPIURL = "https://s2k7tnzlhrpw.statuspage.io/api/v2/summary.json"

// IncidentCollector collects number of active incidents associated with digital ocean services
type IncidentCollector struct {
	logger  log.Logger
	errors  *prometheus.CounterVec
	client  *godo.Client
	timeout time.Duration

	Incidents *prometheus.Desc
}

// NewIncidentCollector returns a new IncidentCollector.
func NewIncidentCollector(logger log.Logger, errors *prometheus.CounterVec, client *godo.Client, timeout time.Duration) *IncidentCollector {
	errors.WithLabelValues("incidents").Add(0)

	return &IncidentCollector{
		logger:  logger,
		errors:  errors,
		client:  client,
		timeout: timeout,

		Incidents: prometheus.NewDesc(
			"digitalocean_incidents",
			"Number of active incidents at digital ocean",
			[]string{},
			nil,
		),
	}
}

// Describe sends the super-set of all possible descriptors of metrics
// collected by this Collector.
func (c *IncidentCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.Incidents
}

// GetIncidents fetches active incidents associated with digital ocean services
func GetIncidents(client *http.Client) ([]gjson.Result, error) {
	r, err := client.Get(doStatusAPIURL)
	if err != nil {
		return []gjson.Result{}, err
	}
	defer r.Body.Close()
	if r.StatusCode == http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return []gjson.Result{}, err
		}
		return gjson.Get(string(bodyBytes), "incidents.#.name").Array(), nil
	}
	return []gjson.Result{}, errors.New("Unable to retrieve incidents")
}

// Collect is called by the Prometheus registry when collecting metrics.
func (c *IncidentCollector) Collect(ch chan<- prometheus.Metric) {
	client := http.Client{Timeout: c.timeout}
	incidents, err := GetIncidents(&client)
	if err != nil {
		c.errors.WithLabelValues("incidents").Add(1)
		level.Warn(c.logger).Log(
			"msg", "can't retrieve incidents",
			"err", err,
		)
	}
	ch <- prometheus.MustNewConstMetric(
		c.Incidents,
		prometheus.GaugeValue,
		float64(len(incidents)),
	)
}
