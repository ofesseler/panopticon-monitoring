package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"strings"

	log "github.com/Sirupsen/logrus"
	api "github.com/ofesseler/panopticon/promapi"
)

var (
	promHost      = flag.String("prom-host", "localhost:9090", "Enter hostname of prometheus")
	listenAddress = flag.String("listen-address", ":8888", "Enter port number to listen on")
	health        = NewHealth("wolke")
)

func main() {
	flag.Parse()

	log.Infof("Start panopticon. Listening on: %v", *listenAddress)

	http.HandleFunc("/", wrapHandler(http.StripPrefix("/", http.FileServer(http.Dir("static")))))

	http.HandleFunc("/api/v1/up", up)
	http.HandleFunc("/api/v1/consul/up", consulUp)
	http.HandleFunc("/api/v1/gluster/up", glusterUp)
	http.HandleFunc("/api/v1/health", healthSummary)
	http.HandleFunc("/api/v1/state/", state)
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

func state(w http.ResponseWriter, r *http.Request) {
	var (
		state    State
		endpoint string
	)
	urlPacks := strings.Split(r.URL.Path, "/")
	endpoint = urlPacks[len(urlPacks)-1]
	state.Last = health.FSM.Current()

	switch endpoint {
	case CURRENT:
		state.Request = CURRENT
		api.FetchHealthSummary(*promHost)
		state.Success = true
	case WARNING:
		state.Request = WARNING
		err := health.FSM.Event(WARNING)
		state.Success = true
		if err != nil {
			log.Error(err)
			state.Message = err.Error()
			state.Success = false
		}

	case FATAL:
		state.Request = FATAL
		err := health.FSM.Event(FATAL)
		state.Success = true
		if err != nil {
			log.Error(err)
			state.Message = err.Error()
			state.Success = false
		}
	case RESOLV:
		state.Request = RESOLV
		err := health.FSM.Event(RESOLV)
		state.Success = true
		if err != nil {
			log.Error(err)
			state.Message = err.Error()
			state.Success = false
		}
	default:
		state.Success = false
		state.Request = fmt.Sprintf("%s %s", r.Method, r.RequestURI)
		state.Message = fmt.Sprintf("Requsted method %s not implemented", endpoint)
	}

	state.Current = health.FSM.Current()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(state)
}
