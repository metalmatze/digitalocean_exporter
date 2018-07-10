{
  prometheusAlerts+:: {
    groups+: [
      {
        name: 'digitalocean',
        rules: [
          {
            expr: |||
              digitalocean_droplet_up == 0
            ||| % $._config,
            labels: {
              severity: 'critical',
            },
            annotations: {
              message: 'Droplet {{ $labels.name }} in region {{ $labels.region }} is down.',
            },
            'for': '5m',
            alert: 'DigitalOceanDropletDown',
          },
          {
            // TODO: Make 100 configurable
            expr: |||
              digitalocean_price_monthly > 100
            ||| % $._config,
            labels: {
              severity: 'warning',
            },
            annotations: {
              message: 'Spending {{ printf "%.2f" $value }}$ at DigitalOcean which exceeds 100$.',
            },
            'for': '1h',
            alert: 'DigitalOceanHighMonthlyPrice',
          },
          {
            expr: |||
              digitalocean_floating_ipv4_active == 0
            ||| % $._config,
            labels: {
              severity: 'warning',
            },
            annotations: {
              message: 'Paying 5$/month for {{ $labels.ipv4 }}, an unused Floating IP.',
            },
            'for': '1h',
            alert: 'DigitalOceanFloatingIPUnused',
          },
          {
            expr: |||
              count(digitalocean_key) < 1
            ||| % $._config,
            labels: {
              severity: 'warning',
            },
            annotations: {
              message: 'We can\'t find SSH keys, please add at least one.',
            },
            'for': '1h',
            alert: 'DigitalOceanNoSSHKeys',
          },
        ],
      },
    ],
  },
}
