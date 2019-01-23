package main

import (
	"fmt"

	hazelcast "github.com/hazelcast/hazelcast-go-client"
	"github.com/prometheus/client_golang/prometheus"
)

var namespace = "hazelcast"

type Exporter struct {
	config *conf

	up             *prometheus.Desc
	scrapeFailures prometheus.Counter
	members        prometheus.Gauge
	maps           *prometheus.GaugeVec
}

func NewExporter(c *conf) *Exporter {
	x := &Exporter{
		config: c,
		up: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "up"),
			"Could the hazelcast server be reached",
			nil,
			nil),
		scrapeFailures: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "exporter_scrape_failures_total",
			Help:      "Number of errors while scraping hazelcast.",
		}),
		members: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "group_members",
			Help:      "Number of members in the group",
		}),
		maps: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "map_items",
			Help:      "Number of items in maps",
		},
			[]string{"map"},
		),
	}
	return x
}

func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.up
	e.scrapeFailures.Describe(ch)
	e.maps.Describe(ch)
}

func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	config := hazelcast.NewConfig()
	config.NetworkConfig().AddAddress(e.config.Url)
	config.GroupConfig().SetName(e.config.ClusterName)
	config.GroupConfig().SetPassword(e.config.ClusterPassword)

	client, err := hazelcast.NewClientWithConfig(config)
	defer client.Shutdown()
	if err != nil {
		ch <- prometheus.MustNewConstMetric(e.up, prometheus.GaugeValue, 0)
		fmt.Println(err)
		return
	}
	ch <- prometheus.MustNewConstMetric(e.up, prometheus.GaugeValue, 1)

	cluster := client.Cluster()
	e.members.Set(float64(len(cluster.GetMembers())))

	for _, mapName := range e.config.Maps {
		m, err := client.GetMap(mapName)
		if err != nil {
			e.maps.WithLabelValues(mapName).Set(-1)
			continue
		}
		size, err := m.Size()
		if err != nil {
			e.maps.WithLabelValues(mapName).Set(-1)
			continue
		}
		e.maps.WithLabelValues(mapName).Set(float64(size))
	}

	e.members.Collect(ch)
	e.maps.Collect(ch)

}
