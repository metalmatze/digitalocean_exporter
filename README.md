# DigitalOcean Exporter [![Build Status](https://drone.github.matthiasloibl.com/api/badges/metalmatze/digitalocean_exporter/status.svg)](https://drone.github.matthiasloibl.com/metalmatze/digitalocean_exporter)

[![Docker Pulls](https://img.shields.io/docker/pulls/metalmatze/digitalocean_exporter.svg?maxAge=604800)](https://hub.docker.com/r/metalmatze/digitalocean_exporter)
[![Go Report Card](https://goreportcard.com/badge/github.com/metalmatze/digitalocean_exporter)](https://goreportcard.com/report/github.com/metalmatze/digitalocean_exporter)

Prometheus exporter for various metrics about your [DigitalOcean](https://www.digitalocean.com/) droplets, volumes, snapshots & networks and much more, written in Go.

### Installation

```bash
go get -u github.com/metalmatze/digitalocean_exporter
```

### Configuration

ENV Variable | Description
|----------|-----|
| WEB_PATH | Path for metrics, default: `/metrics` |
| WEB_ADDR | Address for this exporter to run, default: `:9211` |
| DIGITALOCEAN_TOKEN | Token for API access |

You can get an API token at: https://cloud.digitalocean.com/settings/api/tokens  
Read-only tokens are sufficient.

### Docker

```bash
docker pull metalmatze/digitalocean_exporter
docker run --rm -p 9211:9211 -e DIGITALOCEAN_TOKEN=XXX metalmatze/digitalocean_exporter
```

Example `docker-compose.yml` with Transmission also running in docker.

```yaml
digitalocean_exporter:
    image: metalmatze/digitalocean_exporter
    environment:
    - '-do.token=XXX'
    restart: always
    ports:
    - "127.0.0.1:9211:9211"
```

### Development

```bash
make
```

For development we encourage you to use `make install` instead, it's faster.

Now simply copy the `.env.example` to `.env`, like `cp .env.example .env` and set your preferences.
Now you're good to go.
