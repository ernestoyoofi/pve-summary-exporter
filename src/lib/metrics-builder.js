const client = require("prom-client")

const registry = new client.Registry()
registry.setDefaultLabels({ job: "pve_exporter" })

const pveTimeMetricsGenerateGauge = new client.Gauge({
  name: "pve_time_metrics_generate",
  help: "PVE Time Metrics Generate",
  labelNames: [
    "identify",    // Identify
    "time_format", // Time Format
  ],
  registers: [registry],
})
const pveClusterLabelGauge = new client.Gauge({
  name: "pve_cluster_label",
  help: "PVE Cluster Label",
  labelNames: [
    "identify", // Identify
    "host"
  ],
  registers: [registry],
})
const pveClusterStatusGauge = new client.Gauge({
  name: "pve_cluster_status",
  help: "PVE Cluster Status",
  labelNames: [
    "identify", // Identify
    "is_now",   // Is Now
    "name",     // Name
    "status",   // Status
    "type",     // Type
    "ip",       // IP
    "level",    // Level
  ],
  registers: [registry],
})
const pveClusterLabelGraphNodeGauge = new client.Gauge({
  name: "pve_cluster_label_graph_node",
  help: "PVE Cluster Label Graph Node",
  labelNames: [
    "identify", // Identify
    "label",
  ],
  registers: [registry],
})
const pveClusterLabelStatusNodeGauge = new client.Gauge({
  name: "pve_cluster_label_status_node",
  help: "PVE Cluster Label Status Node",
  labelNames: [
    "identify", // Identify
    "label",
    "value"
  ],
  registers: [registry],
})

function buildingMetrics(dataScrapping = []) {
  registry.resetMetrics()
  for (const dataPVE of dataScrapping) {
    pveTimeMetricsGenerateGauge.set({
      identify: dataPVE.identify,
      time_format: "miliseconds",
    }, dataPVE.timeEnd)
    pveClusterLabelGauge.set({
      identify: dataPVE.identify,
      host: dataPVE.host,
    }, 1)
    if (!!dataPVE?.node?.now) {
      pveClusterStatusGauge.set({
        identify: dataPVE.identify,
        is_now: "true",
        name: dataPVE.node.now.name,
        status: !!dataPVE.node.now.online ? "online" : "offline",
        type: dataPVE.node.now.type,
        ip: dataPVE.node.now.ip,
        level: dataPVE.node.now.level,
      }, 1)
      for (const nodeItem of (dataPVE?.node?.list || [])) {
        pveClusterStatusGauge.set({
          identify: dataPVE.identify,
          is_now: "false",
          name: nodeItem.name,
          status: !!nodeItem.online ? "online" : "offline",
          type: nodeItem.type,
          ip: nodeItem.ip,
          level: nodeItem.level,
        }, 1)
      }
    }
    if (!!dataPVE?.rrddata) {
      for (const keyOfLabel of Object.keys(dataPVE.rrddata)) {
        const valueOfLabel = dataPVE.rrddata[keyOfLabel]
        pveClusterLabelGraphNodeGauge.set({
          identify: dataPVE.identify,
          label: keyOfLabel,
        }, valueOfLabel)
      }
    }
    if (!!dataPVE?.status) {
      for (const keyOfLabel of Object.keys(dataPVE.status)) {
        const valueOfLabel = dataPVE.status[keyOfLabel]
        if (typeof valueOfLabel === "number") {
          pveClusterLabelStatusNodeGauge.set({
            identify: dataPVE.identify,
            label: keyOfLabel,
            value: "",
          }, valueOfLabel)
        }
        if (typeof valueOfLabel === "string") {
          pveClusterLabelStatusNodeGauge.set({
            identify: dataPVE.identify,
            label: keyOfLabel,
            value: String(valueOfLabel),
          }, 1)
        }
      }
    }
  }
}

module.exports = {
  buildingMetrics,
  registry,
}