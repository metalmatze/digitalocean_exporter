package main

import (
	"context"
	"log"
	"net/http"

	arg "github.com/alexflint/go-arg"
	"github.com/digitalocean/godo"
	"github.com/joho/godotenv"
	"github.com/metalmatze/digitalocean_exporter/collector"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/version"
	"golang.org/x/oauth2"
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
	log.Println("Starting digitalocean_exporter", version.Info())

	_ = godotenv.Load()

	c := Config{
		WebPath: "/metrics",
		WebAddr: ":9211",
	}
	arg.MustParse(&c)

	if c.DigitalOceanToken == "" {
		log.Fatal("DigitalOcean Token is required")
	}

	oauthClient := oauth2.NewClient(context.TODO(), c)
	client := godo.NewClient(oauthClient)

	prometheus.MustRegister(collector.NewAccountCollector(client))
	prometheus.MustRegister(collector.NewDropletCollector(client))
	prometheus.MustRegister(collector.NewFloatingIPCollector(client))
	prometheus.MustRegister(collector.NewImageCollector(client))
	prometheus.MustRegister(collector.NewKeyCollector(client))
	prometheus.MustRegister(collector.NewSnapshotCollector(client))
	prometheus.MustRegister(collector.NewVolumeCollector(client))

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

	log.Println("Listening on", c.WebAddr)
	if err := http.ListenAndServe(c.WebAddr, nil); err != nil {
		log.Fatal(err)
	}
}
