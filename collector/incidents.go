package collector

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

const doStatusAPIURL = "https://s2k7tnzlhrpw.statuspage.io/api/v2/summary.json"

var regionRegex = regexp.MustCompile("[A-Z]{3}\\d{1}")

// DOIncidentAPIResponse stores active digitalocean incidents with their Name(title) to extract the region name
type DOIncidentAPIResponse struct {
	Incidents []struct {
		Name string `json:"name"`
	} `json:"incidents"`
}

// IncidentCollector collects number of active incidents associated with digital ocean services
type IncidentCollector struct {
	logger  log.Logger
	errors  *prometheus.CounterVec
	timeout time.Duration

	Incidents      *prometheus.Desc
	IncidentsTotal *prometheus.Desc
}

// NewIncidentCollector returns a new IncidentCollector.
func NewIncidentCollector(logger log.Logger, errors *prometheus.CounterVec, timeout time.Duration) *IncidentCollector {
	errors.WithLabelValues("incidents").Add(0)

	labels := []string{"region"}
	return &IncidentCollector{
		logger:  logger,
		errors:  errors,
		timeout: timeout,

		Incidents: prometheus.NewDesc(
			"digitalocean_incidents",
			"Number of regional active incidents at digitalocean",
			labels, nil,
		),
		IncidentsTotal: prometheus.NewDesc(
			"digitalocean_incidents_total",
			"Number of total active incidents at digitalocean",
			nil, nil,
		),
	}
}

// Describe sends the super-set of all possible descriptors of metrics
// collected by this Collector.
func (c *IncidentCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.Incidents
}

// GetIncidents fetches active incidents associated with digital ocean services
func GetIncidents(client *http.Client) (DOIncidentAPIResponse, error) {
	r, err := client.Get(doStatusAPIURL)
	if err != nil {
		return DOIncidentAPIResponse{}, err
	}
	defer r.Body.Close()

	if r.StatusCode != http.StatusOK {
		return DOIncidentAPIResponse{}, fmt.Errorf("unable to retrieve incidents: %w", err)
	}

	var doIncidents DOIncidentAPIResponse
	if err := json.NewDecoder(r.Body).Decode(&doIncidents); err != nil {
		return DOIncidentAPIResponse{}, err
	}

	return doIncidents, nil
}

// parseRegion extracts the region code for digitalocean datacenters in an incident titl(e.g. NYC1, SFO3)
func parseRegion(s string) string {
	region := regionRegex.FindString(s)
	// Not all incidents have regions reported
	if region == "" {
		return "unspecified"
	}
	return strings.ToLower(region)
}

// Collect is called by the Prometheus registry when collecting metrics.
func (c *IncidentCollector) Collect(ch chan<- prometheus.Metric) {
	// Datastore to count all incidents per region
	regionalIncidents := make(map[string]int)
	client := http.Client{Timeout: c.timeout}
	doStatus, err := GetIncidents(&client)
	if err != nil {
		c.errors.WithLabelValues("incidents").Add(1)
		level.Warn(c.logger).Log(
			"msg", "can't retrieve incidents",
			"err", err,
		)
	}
	// Count all incidents per region
	for _, incident := range doStatus.Incidents {
		// Extract region name from incident title(if present)
		region := parseRegion(incident.Name)
		if _, ok := regionalIncidents[region]; ok {
			// If key is present, increment
			regionalIncidents[region]++
		} else {
			// If key is not present, create with initial value of 1
			regionalIncidents[region] = 1
		}
	}

	// Create metric per region
	for region, incidentCount := range regionalIncidents {
		ch <- prometheus.MustNewConstMetric(
			c.Incidents,
			prometheus.GaugeValue,
			float64(incidentCount),
			region,
		)
	}

	// Create metric for all incidents
	ch <- prometheus.MustNewConstMetric(
		c.IncidentsTotal,
		prometheus.GaugeValue,
		float64(len(doStatus.Incidents)),
	)
}
