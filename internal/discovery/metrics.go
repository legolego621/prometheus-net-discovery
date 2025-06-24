package discovery

import (
	"github.com/prometheus/client_golang/prometheus"
)

const (
	MetricsNamespace = "net_discovery"
)

type Metrics struct {
	DiscoveryRunning       *prometheus.GaugeVec
	DiscoveredHost         *prometheus.Desc
	CollectionCount        *prometheus.CounterVec
	CollectionDurationLast *prometheus.GaugeVec
	CollectionDuration     *prometheus.HistogramVec
}

func NewMetrics(instanceID string) *Metrics {
	constLabels := prometheus.Labels{
		"instance_id": instanceID,
	}

	return &Metrics{
		DiscoveryRunning: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   MetricsNamespace,
				Name:        "discovery_running",
				Help:        "State of running proccess of discovery. 1 if discovery is running, 0 otherwise",
				ConstLabels: constLabels,
			},
			[]string{"network"},
		),
		DiscoveredHost: prometheus.NewDesc(
			prometheus.BuildFQName(MetricsNamespace, "", "discovered_host"),
			"Discovered host ports and icmp",
			[]string{"network", "host", "protocol", "port", "srv"},
			constLabels,
		),
		CollectionCount: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace:   MetricsNamespace,
				Name:        "collection_total",
				Help:        "Count of collections",
				ConstLabels: constLabels,
			},
			[]string{"network"},
		),
		CollectionDurationLast: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   MetricsNamespace,
				Name:        "collection_duration_last",
				Help:        "Duration of last collections in seconds",
				ConstLabels: constLabels,
			},
			[]string{"network"},
		),
		CollectionDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: MetricsNamespace,
				Name:      "collection_duration",
				Help:      "Duration of collections in seconds",
				Buckets: []float64{
					1, 2, 3, 5, 10, 20, 40, 80, 160, 320, 640, 1280,
					2560, 5120, 10240, 20480, 40960, 81920, 163840,
				},
				ConstLabels: constLabels,
			},
			[]string{"network"},
		),
	}
}

func (m *Metrics) Register() {
	prometheus.MustRegister(m.CollectionCount)
	prometheus.MustRegister(m.DiscoveryRunning)
	prometheus.MustRegister(m.CollectionDurationLast)
	prometheus.MustRegister(m.CollectionDuration)
}

func (d *Discovery) MetricsRegister() {
	prometheus.MustRegister(d)
	d.Metrics.Register()
}

func (d *Discovery) Describe(ch chan<- *prometheus.Desc) {
	ch <- d.Metrics.DiscoveredHost
}

func (d *Discovery) Collect(ch chan<- prometheus.Metric) {
	for net, report := range d.Reports {
		for _, h := range report.DiscoveredHosts {
			for _, p := range h.Ports {
				ch <- prometheus.MustNewConstMetric(
					d.Metrics.DiscoveredHost,
					prometheus.GaugeValue,
					1,
					net,
					h.Address,
					p.Protocol,
					p.Port,
					p.Service,
				)
			}

			if h.ICMP {
				ch <- prometheus.MustNewConstMetric(
					d.Metrics.DiscoveredHost,
					prometheus.GaugeValue,
					1,
					net,
					h.Address,
					"icmp",
					"",
					"",
				)
			}
		}
	}
}
