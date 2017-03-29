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
	// Date this binary was built.
	Date string
	// GoVersion running this binary.
	GoVersion = runtime.Version()
	// StartTime has the time this was started.
	StartTime = time.Now()
)

// Config gets its content from env and passes it on to different packages
type Config struct {
	DigitalOceanToken string `arg:"env:DIGITALOCEAN_TOKEN"`
	WebAddr           string `arg:"env:WEB_ADDR"`
	WebPath           string `arg:"env:WEB_PATH"`
}

// Token returns a token or an error.
func (c Config) Token() (*oauth2.Token, error) {
	return &oauth2.Token{AccessToken: c.DigitalOceanToken}, nil
}

func main() {
	_ = godotenv.Load()

	c := Config{
		WebPath: "/metrics",
		WebAddr: ":9212",
	}
	arg.MustParse(&c)

	if c.DigitalOceanToken == "" {
		panic("DigitalOcean Token is required")
	}

	logger := log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
	logger = level.NewFilter(logger, level.AllowDebug())
	logger = log.With(logger,
		"ts", log.DefaultTimestampUTC,
		"caller", log.DefaultCaller,
	)

	level.Info(logger).Log(
		"msg", "starting digitalocean_snapper",
		"version", Version,
		"revision", Revision,
		"date", Date,
		"go", GoVersion,
	)

	oauthClient := oauth2.NewClient(context.TODO(), c)
	client := godo.NewClient(oauthClient)

	prometheus.MustRegister(collector.NewAccountCollector(logger, client))
	prometheus.MustRegister(collector.NewDropletCollector(logger, client))
	prometheus.MustRegister(collector.NewFloatingIPCollector(logger, client))
	prometheus.MustRegister(collector.NewImageCollector(logger, client))
	prometheus.MustRegister(collector.NewKeyCollector(logger, client))
	prometheus.MustRegister(collector.NewSnapshotCollector(logger, client))
	prometheus.MustRegister(collector.NewVolumeCollector(logger, client))

	http.Handle(c.WebPath, promhttp.Handler())
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
