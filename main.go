package main

import (
	"io/ioutil"
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gopkg.in/alecthomas/kingpin.v2"
	yaml "gopkg.in/yaml.v2"
)

type conf struct {
	Url             string   `yaml:"hazelcastUrl"`
	ClusterName     string   `yaml:"hazelcastClusterName"`
	ClusterPassword string   `yaml:"hazelcastClusterPassword"`
	Maps            []string `yaml:"hazelcastMaps"`
}

var (
	addr   = kingpin.Flag("listen", "Address to listen to").Required().String()
	config = kingpin.Flag("config", "Config file").Required().String()
)

func main() {
	kingpin.Parse()
	var c conf
	c.getConf()

	exporter := NewExporter(&c)
	prometheus.MustRegister(exporter)
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(*addr, nil))
}

func (c *conf) getConf() *conf {

	yamlFile, err := ioutil.ReadFile(*config)
	if err != nil {
		log.Fatalf("Read: %v", err)
	}
	err = yaml.UnmarshalStrict(yamlFile, c)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}

	return c
}
