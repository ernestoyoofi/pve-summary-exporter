package metrics

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"pve-summary-exporter/pkg/proxmox"
)

type Metrics struct {
	PveTimeMetricsGenerate      *prometheus.GaugeVec
	PveClusterLabel             *prometheus.GaugeVec
	PveClusterStatus            *prometheus.GaugeVec
	PveClusterLabelGraphNode    *prometheus.GaugeVec
	PveClusterLabelStatusNode   *prometheus.GaugeVec
	Registry                    *prometheus.Registry
}

func NewMetrics() *Metrics {
	m := &Metrics{
		PveTimeMetricsGenerate: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "pve_time_metrics_generate",
			Help: "PVE Time Metrics Generate",
		}, []string{"identify", "time_format"}),
		PveClusterLabel: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "pve_cluster_label",
			Help: "PVE Cluster Label",
		}, []string{"identify", "host"}),
		PveClusterStatus: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "pve_cluster_status",
			Help: "PVE Cluster Status",
		}, []string{"identify", "is_now", "name", "status", "type", "ip", "level"}),
		PveClusterLabelGraphNode: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "pve_cluster_label_graph_node",
			Help: "PVE Cluster Label Graph Node",
		}, []string{"identify", "label"}),
		PveClusterLabelStatusNode: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "pve_cluster_label_status_node",
			Help: "PVE Cluster Label Status Node",
		}, []string{"identify", "label", "value"}),
		Registry: prometheus.NewRegistry(),
	}

	m.Registry.MustRegister(m.PveTimeMetricsGenerate)
	m.Registry.MustRegister(m.PveClusterLabel)
	m.Registry.MustRegister(m.PveClusterStatus)
	m.Registry.MustRegister(m.PveClusterLabelGraphNode)
	m.Registry.MustRegister(m.PveClusterLabelStatusNode)

	return m
}

func (m *Metrics) Reset() {
	m.PveTimeMetricsGenerate.Reset()
	m.PveClusterLabel.Reset()
	m.PveClusterStatus.Reset()
	m.PveClusterLabelGraphNode.Reset()
	m.PveClusterLabelStatusNode.Reset()
}

type ScrapeResult struct {
	Identify string
	Host     string
	Success  bool
	TimeEnd  int64
	Node     *proxmox.ClusterStatusResponse
	RRDData  map[string]float64
	Status   map[string]interface{}
}

func (m *Metrics) BuildMetrics(results []ScrapeResult) {
	m.Reset()
	for _, res := range results {
		m.PveTimeMetricsGenerate.WithLabelValues(res.Identify, "miliseconds").Set(float64(res.TimeEnd))
		m.PveClusterLabel.WithLabelValues(res.Identify, res.Host).Set(1)

		if res.Node != nil {
			if res.Node.Now != nil {
				status := "offline"
				if res.Node.Now.Online == 1 {
					status = "online"
				}
				m.PveClusterStatus.WithLabelValues(
					res.Identify,
					"true",
					res.Node.Now.Name,
					status,
					res.Node.Now.Type,
					res.Node.Now.IP,
					res.Node.Now.Level,
				).Set(1)
			}

			for _, nodeItem := range res.Node.List {
				status := "offline"
				if nodeItem.Online == 1 {
					status = "online"
				}
				m.PveClusterStatus.WithLabelValues(
					res.Identify,
					"false",
					nodeItem.Name,
					status,
					nodeItem.Type,
					nodeItem.IP,
					nodeItem.Level,
				).Set(1)
			}
		}

		if res.RRDData != nil {
			for k, v := range res.RRDData {
				m.PveClusterLabelGraphNode.WithLabelValues(res.Identify, k).Set(v)
			}
		}

		if res.Status != nil {
			for k, v := range res.Status {
				switch val := v.(type) {
				case float64:
					m.PveClusterLabelStatusNode.WithLabelValues(res.Identify, k, "").Set(val)
				case int64:
					m.PveClusterLabelStatusNode.WithLabelValues(res.Identify, k, "").Set(float64(val))
				case int:
					m.PveClusterLabelStatusNode.WithLabelValues(res.Identify, k, "").Set(float64(val))
				case string:
					m.PveClusterLabelStatusNode.WithLabelValues(res.Identify, k, val).Set(1)
				default:
					fmt.Printf("Unknown type for status label %s: %T\n", k, v)
				}
			}
		}
	}
}
