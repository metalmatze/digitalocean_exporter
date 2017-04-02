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