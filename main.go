package main

import (
	"context"
	"flag"
	"log"
	"net/http"

	"github.com/digitalocean/godo"
	"github.com/metalmatze/digitalocean_exporter/collector"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/version"
	"golang.org/x/oauth2"
)

type TokenSource struct {
	AccessToken string
}

func (t *TokenSource) Token() (*oauth2.Token, error) {
	token := &oauth2.Token{
		AccessToken: t.AccessToken,
	}
	return token, nil
}

func main() {
	var (
		listenAddress = flag.String("web.listen-address", ":9100", "Address on which to expose metrics and web interface.")
		metricsPath   = flag.String("web.telemetry-path", "/metrics", "Path under which to expose metrics.")
		token         = flag.String("do.token", "", "Token you got from DigitalOcean for the API (read-only).")
	)
	flag.Parse()

	tokenSource := &TokenSource{
		AccessToken: *token,
	}
	oauthClient := oauth2.NewClient(context.TODO(), tokenSource)
	client := godo.NewClient(oauthClient)

	log.Println("Starting digitalocean_exporter", version.Info())

	prometheus.MustRegister(collector.NewAccountCollector(client))
	prometheus.MustRegister(collector.NewDropletCollector(client))
	prometheus.MustRegister(collector.NewVolumeCollector(client))
	prometheus.MustRegister(collector.NewImageCollector(client))
	prometheus.MustRegister(collector.NewFloatingIPCollector(client))
	prometheus.MustRegister(collector.NewKeyCollector(client))

	http.Handle(*metricsPath, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
			<head><title>DigitalOcean Exporter</title></head>
			<body>
			<h1>DigitalOcean Exporter</h1>
			<p><a href="` + *metricsPath + `">Metrics</a></p>
			</body>
			</html>`))
	})

	log.Println("Listening on", *listenAddress)
	if err := http.ListenAndServe(*listenAddress, nil); err != nil {
		log.Fatal(err)
	}
}
