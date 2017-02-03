package main

import (
	"flag"
	"net/http"

	api "github.com/ofesseler/panopticon/promapi"

	"encoding/json"
	log "github.com/Sirupsen/logrus"
)

var (
	promHost      = flag.String("prom-host", "localhost:9090", "Enter hostname of prometheus")
	listenAddress = flag.String("listen-address", ":8888", "Enter port number to listen on")
)

func main() {

	flag.Parse()

	log.Infof("Start panopticon. Listening on: %v", *listenAddress)

	http.HandleFunc("/", wrapHandler(http.StripPrefix("/", http.FileServer(http.Dir("static")))))

	http.HandleFunc("/api/v1/up", up)
	http.HandleFunc("/api/v1/consul/up", consulUp)
	http.HandleFunc("/api/v1/gluster/up", glusterUp)
	http.HandleFunc("/api/v1/health", healthSummary)

	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}

func up(w http.ResponseWriter, r *http.Request) {

	upHealthStatus, err := api.CheckUp(*promHost, api.Up)
	if err != nil {
		log.Error(err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(upHealthStatus)
}

func consulUp(w http.ResponseWriter, r *http.Request) {

	consulUpHealthStatus, err := api.CheckUp(*promHost, api.ConsulUp)
	if err != nil {
		log.Error(err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(consulUpHealthStatus)
}

func glusterUp(w http.ResponseWriter, r *http.Request) {

	glusterUpHealthStatus, err := api.CheckUp(*promHost, api.GlusterUp)
	if err != nil {
		log.Error(err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(glusterUpHealthStatus)
}

func healthSummary(w http.ResponseWriter, r *http.Request) {

	healthSummary, err := api.FetchHealthSummary(*promHost)
	if err != nil {
		log.Error(err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(healthSummary)
}
