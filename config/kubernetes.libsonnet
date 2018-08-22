local k = import 'ksonnet/ksonnet.beta.3/k.libsonnet';

{
  _config+:: {
    namespace: 'monitoring',

    versions+:: {
      digitalOceanExporter: '0.5',
    },

    imageRepos+:: {
      digitalOceanExporter: 'metalmatze/digitalocean_exporter',
    },

    digitalOceanExporter+:: {
      port: 9212,
    },
  },
  digitalOceanExporter+: {
    serviceMonitor: {
      apiVersion: 'monitoring.coreos.com/v1',
      kind: 'ServiceMonitor',
      metadata: {
        name: 'digitalocean-exporter',
        namespace: $._config.namespace,
        labels: {
          'k8s-app': 'digitalocean-exporter',
        },
      },
      spec: {
        jobLabel: 'k8s-app',
        selector: {
          matchLabels: $.digitalOceanExporter.deployment.spec.selector.matchLabels,
        },
        endpoints: [
          {
            port: 'http',
            interval: '10m',
          },
        ],
      },
    },
    service:
      local service = k.core.v1.service;
      local servicePort = k.core.v1.service.mixin.spec.portsType;

      service.new(
        'digitalocean-exporter',
        $.digitalOceanExporter.deployment.spec.selector.matchLabels,
        servicePort.newNamed('http', $._config.digitalOceanExporter.port, 'http'),
      ) +
      service.mixin.metadata.withLabels($.digitalOceanExporter.deployment.spec.selector.matchLabels) +
      service.mixin.metadata.withNamespace($._config.namespace),
    deployment:
      local deployment = k.apps.v1beta2.deployment;
      local container = k.apps.v1beta2.deployment.mixin.spec.template.spec.containersType;
      local containerPort = container.portsType;
      local containerEnv = container.envType;

      local podLabels = { app: 'digitalocean-exporter' };

      local c =
        container.new('digitalocean-exporter', $._config.imageRepos.digitalOceanExporter + ':' + $._config.versions.digitalOceanExporter) +
        container.withPorts(containerPort.newNamed('http', $._config.digitalOceanExporter.port)) +
        container.withEnv([
          containerEnv.fromSecretRef('DIGITALOCEAN_TOKEN', 'digitalocean-exporter', 'DIGITALOCEAN_TOKEN'),
        ]) +
        container.mixin.resources.withRequests({ cpu: '50m', memory: '32Mi' }) +
        container.mixin.resources.withLimits({ cpu: '100m', memory: '128Mi' });

      deployment.new('digitalocean-exporter', 1, c, podLabels) +
      deployment.mixin.metadata.withNamespace($._config.namespace) +
      deployment.mixin.metadata.withLabels(podLabels) +
      deployment.mixin.spec.selector.withMatchLabels(podLabels) +
      deployment.mixin.spec.template.spec.securityContext.withRunAsNonRoot(true) +
      deployment.mixin.spec.template.spec.securityContext.withRunAsUser(65534),
  },
}
