{
  prometheusRules+:: {
    groups+: [
      {
        name: 'digitalocean.rules',
        rules: [
          {
            record: 'digitalocean_droplets_price_monthly',
            expr: |||
              sum(digitalocean_droplet_price_monthly)
            ||| % $._config,
          },
          {
            record: 'digitalocean_snapshots_price_monthly',
            expr: |||
              sum(digitalocean_snapshot_size_bytes) / 1024^3 / 20
            ||| % $._config,
          },
          {
            record: 'digitalocean_volumes_price_monthly',
            expr: |||
              sum(digitalocean_volume_size_bytes) / 1024^3 / 10
            ||| % $._config,
          },
          {
            record: 'digitalocean_price_monthly',
            expr: |||
              digitalocean_droplets_price_monthly +
              digitalocean_volumes_price_monthly +
              digitalocean_snapshots_price_monthly
            ||| % $._config,
          },
        ],
      },
    ],
  },
}
