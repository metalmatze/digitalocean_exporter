# DigitalOcean Exporter [![Build Status](https://cloud.drone.io/api/badges/metalmatze/digitalocean_exporter/status.svg)](https://cloud.drone.io/metalmatze/digitalocean_exporter)

[![Docker Pulls](https://img.shields.io/docker/pulls/metalmatze/digitalocean_exporter.svg?maxAge=604800)](https://hub.docker.com/r/metalmatze/digitalocean_exporter)
[![Go Report Card](https://goreportcard.com/badge/github.com/metalmatze/digitalocean_exporter)](https://goreportcard.com/report/github.com/metalmatze/digitalocean_exporter)

Prometheus exporter for various metrics about your [DigitalOcean](https://www.digitalocean.com/) droplets, volumes, snapshots & networks and much more, written in Go.

### Installation

For pre-built binaries please take a look at the releases.  
https://github.com/metalmatze/digitalocean_exporter/releases

To deploy directly onto digitalocean, click the button below.  
[![Deploy to DO](https://mp-assets1.sfo2.digitaloceanspaces.com/deploy-to-do/do-btn-blue.svg)](https://cloud.digitalocean.com/apps/new?repo=https://github.com/metalmatze/digitalocean_exporter/tree/master)

#### Docker

```bash
docker pull metalmatze/digitalocean_exporter:0.6.1
docker run --rm -p 9212:9212 -e DIGITALOCEAN_TOKEN=XXX metalmatze/digitalocean_exporter:0.6.1
```

Example `docker-compose.yml` with Transmission also running in docker.

```yaml
digitalocean_exporter:
    image: metalmatze/digitalocean_exporter:0.6.1
    environment:
    - '-do.token=XXX'
    restart: always
    ports:
    - "127.0.0.1:9212:9212"
```

### Configuration

| ENV Variable                          | Description                                                               |
|---------------------------------------|---------------------------------------------------------------------------|
| DEBUG                                 | If set to true also debug information will be logged, otherwise only info |
| DIGITALOCEAN_TOKEN                    | Token for API access                                                      |
| DIGITALOCEAN_SPACES_ACCESS_KEY_ID     | Spaces Access Key ID to list buckets                                      |
| DIGITALOCEAN_SPACES_ACCESS_KEY_SECRET | Spaces Access Key Secret to list buckets                                  |
| HTTP_TIMEOUT                          | Timeout for the godo client, default: `5000`ms                            |
| WEB_ADDR                              | Address for this exporter to run, default: `:9212`                        |
| WEB_PATH                              | Path for metrics, default: `/metrics`                                     |

You can get an API token at: https://cloud.digitalocean.com/settings/api/tokens  
Read-only tokens are sufficient.

### Metrics

|Name                                         |Type     |Cardinality   |Help
|----                                         |----     |-----------   |----
| digitalocean_account_active                 | gauge   | 1            | The status of your account
| digitalocean_account_balance                | gauge   | 1            | Current balance of your most recent billing activity
| digitalocean_account_droplet_limit          | gauge   | 1            | The maximum number of droplet you can use
| digitalocean_account_floating_ip_limit      | gauge   | 1            | The maximum number of floating ips you can use
| digitalocean_account_verified               | gauge   | 1            | 1 if your email address was verified
| digitalocean_app                            | gauge   | 5            | A metric with a constant '1' value labeled by app id, name, tier, region, and app phase("BUILDING", "DEPLOYING", "ACTIVE", "SUPERSEDED")
| digitalocean_balance_generated_at           | gauge   | 1            | The time at which balances were most recently generated
| digitalocean_build_info                     | gauge   | 1            | A metric with a constant '1' value labeled by version, revision, and branch from which the node_exporter was built.
| digitalocean_database_status                | gauge   | 9            | The status of the database, 1 if online, 0 otherwise
| digitalocean_database_nodes                 | gauge   | 9            | The number of nodes in a database cluster
| digitalocean_domain_record_port             | gauge   | 7            | The port for SRV records
| digitalocean_domain_record_priority         | gauge   | 7            | The priority for SRV and MX records
| digitalocean_domain_record_weight           | gauge   | 7            | The weight for SRV records
| digitalocean_domain_ttl_seconds             | gauge   | 1            | Seconds that clients can cache queried information before a refresh should be requested
| digitalocean_droplet_cpus                   | gauge   | 4            | Droplet's number of CPUs
| digitalocean_droplet_disk_bytes             | gauge   | 4            | Droplet's disk in bytes
| digitalocean_droplet_memory_bytes           | gauge   | 4            | Droplet's memory in bytes
| digitalocean_droplet_price_hourly           | gauge   | 4            | Price of the Droplet billed hourly in dollars
| digitalocean_droplet_price_monthly          | gauge   | 4            | Price of the Droplet billed monthly in dollars
| digitalocean_droplet_up                     | gauge   | 4            | If 1 the droplet is up and running, 0 otherwise
| digitalocean_floating_ipv4_active           | gauge   | 1            | If 1 the floating ip used by a droplet, 0 otherwise
| digitalocean_incidents                      | gauge   | 1            | Number of active regional incidents associated with digitalocean services
| digitalocean_incidents_total                | gauge   | 0            | Number of active total incidents associated with digitalocean services
| digitalocean_key                            | gauge   | 1            | Information about keys in your digitalocean account
| digitalocean_loadbalancer_droplets          | gauge   | 1            | The number of droplets this load balancer is proxying to
| digitalocean_loadbalancer_status            | gauge   | 1            | The status of the load balancer, 1 if active
| digitalocean_month_to_date_balance          | gauge   | 1            | Balance as of the `digitalocean_balance_generated_at` time
| digitalocean_month_to_date_usage            | gauge   | 1            | Amount used in the current billing period as of the `digitalocean_balance_generated_at` time
| digitalocean_snapshot_min_disk_size_bytes   | gauge   | 2            | Minimum disk size for a droplet/volume to run this snapshot on in bytes
| digitalocean_snapshot_size_bytes            | gauge   | 2            | Snapshot's size in bytes
| digitalocean_spaces_bucket                  | gauge   | 2            | Spaces bucket, will always be 1. Includes name and region labels
| digitalocean_spaces_bucket_created          | gauge   | 2            | Spaces bucket creation timestamp in unix epoch format. Includes name and region labels
| digitalocean_start_time                     | gauge   | 1            | Unix timestamp of the start time
| digitalocean_volume_size_bytes              | gauge   | 11           | Volume's size in bytes

### Alerts & Recording Rules

As example alerts and recording rules I have copied my `.rules` file to this repository.  
Please check [example.rules.yaml](example.rules.yml).

### Development

You obviously should get the code

```bash
go get -u github.com/metalmatze/digitalocean_exporter
```

This should already put a binary called `digitalocean_exporter` into `$GOPATH/bin`.

Make sure you copy the `.env.example` to `.env` and change this one to your preferences.

Now during development I always run:

```bash
make install && digitalocean_exporter
```

Use `make install` which uses `go install` in the background to build faster during development.
