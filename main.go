package main

import (
	"context"
	"net/http"
	"os"
	"runtime"
	"time"

	arg "github.com/alexflint/go-arg"
	"github.com/digitalocean/godo"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/joho/godotenv"
	"github.com/metalmatze/digitalocean_exporter/collector"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/oauth2"
)

var (
	// Version of digitalocean_exporter.
	Version string
	// Revision or Commit this binary was built from.
	Revision string
	// BuildDate this binary was built.
	BuildDate string
	// GoVersion running this binary.
	GoVersion = runtime.Version()
	// StartTime has the time this was started.
	StartTime = time.Now()
)

// Config gets its content from env and passes it on to different packages
type Config struct {
	Debug                 bool   `arg:"env:DEBUG"`
	DigitalOceanToken     string `arg:"env:DIGITALOCEAN_TOKEN"`
	SpacesAccessKeyID     string `arg:"env:DIGITALOCEAN_SPACES_ACCESS_KEY_ID"`
	SpacesAccessKeySecret string `arg:"env:DIGITALOCEAN_SPACES_ACCESS_KEY_SECRET"`
	HTTPTimeout           int    `arg:"env:HTTP_TIMEOUT"`
	WebAddr               string `arg:"env:WEB_ADDR"`
	WebPath               string `arg:"env:WEB_PATH"`
}

// Token returns a token or an error.
func (c Config) Token() (*oauth2.Token, error) {
	return &oauth2.Token{AccessToken: c.DigitalOceanToken}, nil
}

func main() {
	_ = godotenv.Load()

	c := Config{
		HTTPTimeout: 5000,
		WebPath:     "/metrics",
		WebAddr:     ":9212",
	}
	arg.MustParse(&c)

	if c.DigitalOceanToken == "" {
		panic("DigitalOcean Token is required")
	}

	filterOption := level.AllowInfo()
	if c.Debug {
		filterOption = level.AllowDebug()
	}

	logger := log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
	logger = level.NewFilter(logger, filterOption)
	logger = log.With(logger,
		"ts", log.DefaultTimestampUTC,
		"caller", log.DefaultCaller,
	)

	level.Info(logger).Log(
		"msg", "starting digitalocean_exporter",
		"version", Version,
		"revision", Revision,
		"buildDate", BuildDate,
		"goVersion", GoVersion,
	)

	if c.SpacesAccessKeyID == "" && c.SpacesAccessKeySecret == "" {
		level.Warn(logger).Log(
			"msg", "Spaces Access Key ID and Secret unset. Spaces buckets will not be collected",
		)
	}

	oauthClient := oauth2.NewClient(context.TODO(), c)
	client := godo.NewClient(oauthClient)

	timeout := time.Duration(c.HTTPTimeout) * time.Millisecond

	errors := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "digitalocean_errors_total",
		Help: "The total number of errors per collector",
	}, []string{"collector"})

	r := prometheus.NewRegistry()
	r.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
	r.MustRegister(prometheus.NewGoCollector())
	r.MustRegister(errors)
	r.MustRegister(collector.NewExporterCollector(logger, Version, Revision, BuildDate, GoVersion, StartTime))
	r.MustRegister(collector.NewAccountCollector(logger, errors, client, timeout))
	r.MustRegister(collector.NewAppCollector(logger, errors, client, timeout))
	r.MustRegister(collector.NewBalanceCollector(logger, errors, client, timeout))
	r.MustRegister(collector.NewDBCollector(logger, errors, client, timeout))
	r.MustRegister(collector.NewDomainCollector(logger, errors, client, timeout))
	r.MustRegister(collector.NewDropletCollector(logger, errors, client, timeout))
	r.MustRegister(collector.NewFloatingIPCollector(logger, errors, client, timeout))
	r.MustRegister(collector.NewImageCollector(logger, errors, client, timeout))
	r.MustRegister(collector.NewKeyCollector(logger, errors, client, timeout))
	r.MustRegister(collector.NewLoadBalancerCollector(logger, errors, client, timeout))
	r.MustRegister(collector.NewSnapshotCollector(logger, errors, client, timeout))
	r.MustRegister(collector.NewVolumeCollector(logger, errors, client, timeout))
	r.MustRegister(collector.NewKubernetesCollector(logger, errors, client, timeout))
	r.MustRegister(collector.NewMonitoringCollector(logger, errors, client, timeout))

	// Only run spaces bucket collector if access key id and secret are set
	if c.SpacesAccessKeyID != "" && c.SpacesAccessKeySecret != "" {
		r.MustRegister(collector.NewSpacesCollector(logger, errors, client, c.SpacesAccessKeyID, c.SpacesAccessKeySecret, timeout))
	}

	http.Handle(c.WebPath,
		promhttp.HandlerFor(r, promhttp.HandlerOpts{}),
	)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`<html>
			<head><title>DigitalOcean Exporter</title></head>
			<body>
			<h1>DigitalOcean Exporter</h1>
			<p><a href="` + c.WebPath + `">Metrics</a></p>
			</body>
			</html>`))
	})

	level.Info(logger).Log("msg", "listening", "addr", c.WebAddr)
	if err := http.ListenAndServe(c.WebAddr, nil); err != nil {
		level.Error(logger).Log("msg", "http listenandserve error", "err", err)
		os.Exit(1)
	}
}
