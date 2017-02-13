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
	health        = NewHealth("wolke")
	conn connection
)

type connection struct {
	*api.PrometheusFetcher
	PromHost string
}

func main() {
	var (
		promHostFlag = flag.String("prom-host", "localhost:9090", "Enter hostname of prometheus")
		listenAddress = flag.String("listen-address", ":8888", "Enter port number to listen on")
	)
	flag.Parse()

	conn.PromHost = *promHostFlag
	conn.PrometheusFetcher = new(api.PrometheusFetcher)

	log.Infof("Start panopticon. Listening on: %v", *listenAddress)

	http.HandleFunc("/", wrapHandler(http.StripPrefix("/", http.FileServer(http.Dir("static")))))

	http.HandleFunc("/api/v1/up", up)
	http.HandleFunc("/api/v1/consul/up", consulUp)
	http.HandleFunc("/api/v1/consul/health", consulHealth)
	http.HandleFunc("/api/v1/gluster/up", glusterUp)
	http.HandleFunc("/api/v1/gluster/health", glusterHealth)
	http.HandleFunc("/api/v1/weave/health", weaveHealth)

	http.HandleFunc("/api/v1/health", healthSummary)
	http.HandleFunc("/api/v1/state/", state)
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}

func up(w http.ResponseWriter, r *http.Request) {
	var httpFetch api.Fetcher = api.PrometheusFetcher{}
	upHealthStatus, err := api.FetchServiceUp(httpFetch, api.Up, conn.PromHost)
	if err != nil {
		log.Error(err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(upHealthStatus)
}

func consulUp(w http.ResponseWriter, r *http.Request) {
	var httpFetch api.Fetcher = api.PrometheusFetcher{}
	consulUpHealthStatus, err := api.FetchServiceUp(httpFetch, api.ConsulUp, conn.PromHost)
	if err != nil {
		log.Error(err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(consulUpHealthStatus)
}

func consulHealth(w http.ResponseWriter, r *http.Request) {
	var f api.Fetcher = api.PrometheusFetcher{}
	h, err := ProcessConsulHealthSummary(f, conn.PromHost)
	if err != nil {
		log.Error(err)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(h)
}

func glusterUp(w http.ResponseWriter, r *http.Request) {
	var httpFetch api.Fetcher = api.PrometheusFetcher{}
	glusterUpHealthStatus, err := api.FetchServiceUp(httpFetch, conn.PromHost, api.GlusterUp)
	if err != nil {
		log.Error(err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(glusterUpHealthStatus)
}

func glusterHealth(w http.ResponseWriter, r *http.Request) {
	var f api.Fetcher = api.PrometheusFetcher{}
	h, err := ProcessGlusterHealthSummary(f, conn.PromHost)
	if err != nil {
		log.Error(err)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(h)
}

func weaveHealth(w http.ResponseWriter, r *http.Request) {
	var f api.Fetcher = api.PrometheusFetcher{}
	h, err := ProcessWeaveHealthSummary(f, conn.PromHost)
	if err != nil {
		log.Error(err)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(h)
}

func healthSummary(w http.ResponseWriter, r *http.Request) {
	var f api.Fetcher = api.PrometheusFetcher{}
	hs, err := ProcessHealthSummary(f, conn.PromHost)
	if err != nil {
		log.Error(err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(hs)
}

func state(w http.ResponseWriter, r *http.Request) {
	var (
		state    State
		endpoint string
		f        api.Fetcher = api.PrometheusFetcher{}
	)
	urlPacks := strings.Split(r.URL.Path, "/")
	endpoint = urlPacks[len(urlPacks)-1]
	state.Last = health.FSM.Current()

	switch endpoint {
	case CURRENT:
		state.Request = CURRENT
		summary, err := api.FetchHealthSummary(f, conn.PromHost)
		if err != nil {
			state.Success = false
			state.Message = err.Error()
		}
		state.Success = summary.Status
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
