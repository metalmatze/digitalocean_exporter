local grafana = import 'grafonnet/grafana.libsonnet';
local dashboard = grafana.dashboard;
local graphPanel = grafana.graphPanel;
local prometheus = grafana.prometheus;
local row = grafana.row;
local singlestat = grafana.singlestat;
local template = grafana.template;

{
  grafanaDashboards+:: {
    'digitalocean.json':
      local monthlyTotalVATEUR = singlestat.new(
        'Monthly Price incl. VAT',
        datasource='$datasource',
        format='currencyEUR',
        valueName='current',
        span=2,
      ).addTarget(prometheus.target(
        '(digitalocean_price_monthly + (digitalocean_price_monthly * 0.19)) * 0.85' % $._config,
      ));

      local monthlyTotalVAT = singlestat.new(
        'Monthly Price incl. VAT',
        datasource='$datasource',
        format='currencyUSD',
        valueName='current',
        span=2,
      ).addTarget(prometheus.target(
        'digitalocean_price_monthly + digitalocean_price_monthly * 0.19' % $._config,
      ));

      local monthlyTotal = singlestat.new(
        'Monthly Price',
        datasource='$datasource',
        format='currencyUSD',
        valueName='current',
        span=2,
      ).addTarget(prometheus.target(
        'digitalocean_price_monthly' % $._config,
      ));

      local monthlyDropletTotal = singlestat.new(
        'Monthly Droplet Price',
        datasource='$datasource',
        format='currencyUSD',
        valueName='current',
        span=2,
      ).addTarget(prometheus.target(
        'digitalocean_droplets_price_monthly' % $._config,
      ));

      local monthlyVolumeTotal = singlestat.new(
        'Monthly Volumes Price',
        datasource='$datasource',
        format='currencyUSD',
        valueName='current',
        span=2,
      ).addTarget(prometheus.target(
        'digitalocean_volumes_price_monthly' % $._config,
      ));

      local monthlySnapshotTotal = singlestat.new(
        'Monthly Snapshots Price',
        datasource='$datasource',
        format='currencyUSD',
        valueName='current',
        span=2,
      ).addTarget(prometheus.target(
        'digitalocean_snapshots_price_monthly' % $._config,
      ));

      local pricingRow =
        row.new()
        .addPanel(monthlyTotalVATEUR)
        .addPanel(monthlyTotalVAT)
        .addPanel(monthlyTotal)
        .addPanel(monthlyDropletTotal)
        .addPanel(monthlyVolumeTotal)
        .addPanel(monthlySnapshotTotal);

      local dropletsTotal = singlestat.new(
        'Droplets',
        datasource='$datasource',
        valueName='current',
        span=3,
      ).addTarget(prometheus.target(
        'count(digitalocean_droplet_up)' % $._config,
      ));

      local dropletsCPUsTotal = singlestat.new(
        'Total CPUs',
        datasource='$datasource',
        valueName='current',
        span=3,
      ).addTarget(prometheus.target(
        'sum(digitalocean_droplet_cpus)' % $._config,
      ));

      local dropletsMemoryTotal = singlestat.new(
        'Total Memory',
        datasource='$datasource',
        format='bytes',
        valueName='current',
        span=3,
      ).addTarget(prometheus.target(
        'sum(digitalocean_droplet_memory_bytes)' % $._config,
      ));

      local dropletsSizeTotal = singlestat.new(
        'Total Disk Size',
        datasource='$datasource',
        format='decbytes',
        valueName='current',
        span=3,
      ).addTarget(prometheus.target(
        'sum(digitalocean_droplet_disk_bytes)' % $._config,
      ));

      local dropletsRow =
        row.new()
        .addPanel(dropletsTotal)
        .addPanel(dropletsCPUsTotal)
        .addPanel(dropletsMemoryTotal)
        .addPanel(dropletsSizeTotal);

      local volumesRow =
        row.new()
        .addPanel(
          singlestat.new(
            'Volumes',
            datasource='$datasource',
            valueName='current',
          ).addTarget(prometheus.target(
            'count(digitalocean_volume_size_bytes)' % $._config,
          )),
        )
        .addPanel(
          singlestat.new(
            'Total Volumes Size',
            datasource='$datasource',
            format='bytes',
            valueName='current',
          ).addTarget(prometheus.target(
            'sum(digitalocean_volume_size_bytes)' % $._config,
          ))
        )
        .addPanel(
          singlestat.new(
            'Total Volume Snapshots Size',
            datasource='$datasource',
            format='bytes',
            valueName='current',
          ).addTarget(prometheus.target(
            'sum(digitalocean_snapshot_size_bytes{type="volume"})'
          ))
        );

      dashboard.new(
        'DigitalOcean',
        time_from='now-1h',
      ).addTemplate(
        {
          current: {
            text: 'Prometheus',
            value: 'Prometheus',
          },
          hide: 0,
          label: null,
          name: 'datasource',
          options: [],
          query: 'prometheus',
          refresh: 1,
          regex: '',
          type: 'datasource',
        },
      )
      .addRow(pricingRow)
      .addRow(dropletsRow)
      .addRow(volumesRow),
  },
}
