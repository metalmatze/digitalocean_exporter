## 0.6.1 / 2020-09-29

* [ENHANCEMENT] Iterate through droplet pages [#12]

## 0.6.0 / 2020-08-06

* [FEATURE] Add Kubernetes Collector [#5]
* [FEATURE] Add Incidents Collector - exposes Gauge about number of ongoing incidents [#9]

## 0.5 / 2018-04-05

* [BREAKING] Update dependencies and use dep over govendor [9b6c995]
* [FEATURE] Add a LoadBalancer Collector [4#]
* [ENHANCEMENT] Convert the example rules to prom v2 yaml [a9de2b8]

## 0.4 / 2017-04-03

* [FEATURE] Add metrics for `domain_record_port`, `domain_record_priority`, `domain_record_weight`, `domain_ttl_seconds`.
* [ENHANCEMENT] Use timeouts with a default 5000ms for evey godo request.

## 0.3 / 2017-03-30

* [FEATURE] Add example.rules with my own recording rules & alerts as examples.
* [FEATURE] Add DEBUG env variable to change log filter on the server.
* [FEATURE] Expose build_info & start_time metric.
* [ENHANCEMENT] Use go-kit logger everywhere with levels.
* [BUGFIX] Don’t crash if floating ip not assigned to droplet. [#1]

## 0.2 / 2017-03-25

* [ENHANCEMENT] Rename `digitalocean_account_status` to `digitalocean_account_active` 
with  a simple 0 or 1 for false and true.
* [ENHANCEMENT] Drone CI now builds docker images only with version tags, no more latest.
* [ENHANCEMENT] Drone CI cross compiles binaries and uploads them to GitHub.

## 0.1 / 2017-03-24

Initial release.
