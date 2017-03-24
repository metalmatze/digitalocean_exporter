# DigitalOcean Exporter [![Build Status](https://drone.github.matthiasloibl.com/api/badges/metalmatze/digitalocean_exporter/status.svg)](https://drone.github.matthiasloibl.com/metalmatze/digitalocean_exporter)

[![Docker Pulls](https://img.shields.io/docker/pulls/metalmatze/digitalocean_exporter.svg?maxAge=604800)](https://hub.docker.com/r/metalmatze/digitalocean_exporter)
[![Go Report Card](https://goreportcard.com/badge/github.com/metalmatze/digitalocean_exporter)](https://goreportcard.com/report/github.com/metalmatze/digitalocean_exporter)

Prometheus exporter for various metrics about your [DigitalOcean](https://www.digitalocean.com/) droplets, volumes, snapshots & networks and much more, written in Go.

### Installation

```bash
go get -u github.com/metalmatze/digitalocean_exporter
```

#### Docker

```bash
docker pull metalmatze/digitalocean_exporter
docker run --rm -p 9212:9212 -e DIGITALOCEAN_TOKEN=XXX metalmatze/digitalocean_exporter
```

Example `docker-compose.yml` with Transmission also running in docker.

```yaml
digitalocean_exporter:
    image: metalmatze/digitalocean_exporter
    environment:
    - '-do.token=XXX'
    restart: always
    ports:
    - "127.0.0.1:9212:9212"
```

### Configuration

ENV Variable | Description
|----------|-----|
| WEB_PATH | Path for metrics, default: `/metrics` |
| WEB_ADDR | Address for this exporter to run, default: `:9212` |
| DIGITALOCEAN_TOKEN | Token for API access |

You can get an API token at: https://cloud.digitalocean.com/settings/api/tokens  
Read-only tokens are sufficient.

### Metrics

All metrics have a prefix `digitalocean_` which is omitted in this overview.

| Metric | Type | Help |
| -------|------|------|
| account_droplet_limit | gauge | The maximum number of droplet you can use |
| account_floating_ip_limit | gauge | The maximum number of floating ips you can use |
| account_status | gauge | The status of your account |
| account_verified | gauge | 1 if your email address was verified |
| droplet_cpus | gauge | Droplet's number of CPUs |
| droplet_disk_bytes | gauge | Droplet's disk in bytes |
| droplet_memory_bytes | gauge | Droplet's memory in bytes |
| droplet_price_hourly | gauge | Price of the Droplet billed hourly |
| droplet_price_monthly | gauge | Price of the Droplet billed monthly |
| droplet_up | gauge | If 1 the droplet is up and running, 0 otherwise |
| floating_ipv4_active | gauge | If 1 the floating ip used by a droplet, 0 otherwise |
| image_min_disk_size_bytes | gauge | Minimum disk size for a droplet to run this image on in bytes |
| key | gauge | Information about keys in your digitalocean account |
| snapshot_min_disk_size_bytes | gauge | Minimum disk size for a droplet/volume to run this snapshot on in bytes |
| snapshot_size_bytes | gauge | Snapshot's size in bytes |
| volume_size_bytes | gauge | Volume's size in bytes |

### Development

```bash
make
```

For development we encourage you to use `make install` instead, it's faster.

Now simply copy the `.env.example` to `.env`, like `cp .env.example .env` and set your preferences.
Now you're good to go.
