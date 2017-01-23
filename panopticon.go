package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"time"

	log "github.com/Sirupsen/logrus"

)


var (
	promHost      = flag.String("prom-host", "localhost:9090", "Enter hostname of prometheus")
	listenAddress = flag.String("listen-address", ":8888", "Enter port number to listen on")
)

// StatusCheckReceived struct represents json response
type StatusCheckReceived struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Metric struct {
				Name     string `json:"__name__"`
				Check    string `json:"check"`
				Instance string `json:"instance"`
				Job      string `json:"job"`
				Node     string `json:"node"`
			} `json:"metric"`
			Value []interface{} `json:"value"`
		} `json:"result"`
	} `json:"data"`
}

// ErrorStatus struct represents json error response
type ErrorStatus struct {
	Status    string `json:"status"`
	ErrorType string `json:"errorType"`
	Error     string `json:"error"`
}

// Node struct represents json node response
type Node struct {
	Instance string `json:"instance"`
	Group    int    `json:"group"`
	ID       int    `json:"id"`
}

// Link struct represents json
type Link struct {
	Source string `json:"source"`
	Target string `json:"target"`
	Value  int    `json:"value"`
}

// Status sturct represents json
type Status struct {
	Nodes []Node `json:"nodes"`
	Links []Link `json:"links"`
}

func checkerr(err error) {
	if err != nil {
		log.WithFields(log.Fields{"checkerr": "err"}).Error(err)
	}
}

func checkPromResponse(resp StatusCheckReceived) bool {
	if resp.Status != "success" {
		log.WithFields(log.Fields{"response": resp}).Error("prometheus request failed")
		return false
	}
	if resp.Data.ResultType != "vector" {
		log.WithFields(log.Fields{"ResultType": resp.Data.ResultType}).Error("prometheus request returned wrong ResultType (other than 'vector')")
		return false
	}
	return true
}

func promQuery(query string) (StatusCheckReceived, ErrorStatus) {
	var errorStatus ErrorStatus

	apiURL := fmt.Sprintf("http://%v/api/v1/query", *promHost)
	urlValues := url.Values{}
	urlValues.Set("query", query)
	concatedURL := apiURL + "?" + urlValues.Encode()
	checkURL, err := url.Parse(concatedURL)
	checkerr(err)
	request, err := http.NewRequest("GET", checkURL.String(), nil)
	checkerr(err)
	client := &http.Client{
		Timeout: 3 * time.Second,
	}
	resp, err := client.Do(request)

	checkerr(err)
	defer resp.Body.Close()
	body, _ := decodeResponse(resp)

	return body, errorStatus
}

func decodeResponse(response *http.Response) (StatusCheckReceived, ErrorStatus) {
	var errorStatus ErrorStatus
	decoder := json.NewDecoder(response.Body)
	var body StatusCheckReceived
	err := decoder.Decode(&body)
	checkerr(err)
	if body.Status != "success" {
		log.Fatal("Prometheus responded with error:", body.Status)
	}
	return body, errorStatus
}

func index(w http.ResponseWriter, r *http.Request) {
	status := make(map[string]string)
	status["status"] = "ok"
	status["health"] = "green"
	json.NewEncoder(w).Encode(status)
}

func upFunc() ([]Node, []Link) {
	resp, _ := promQuery(`up{job="consul_wolke"}`)
	checkPromResponse(resp)
	var nodesStatus []Node
	var linksStatus []Link
	for _, v := range resp.Data.Result {
		value, _ := strconv.Atoi(v.Value[1].(string))
		node := Node{Instance: v.Metric.Instance, Group: value}
		nodesStatus = append(nodesStatus, node)
	}
	linksStatus = getLinksForNodes(nodesStatus)
	return nodesStatus, linksStatus
}

func getLinksForNodes(nodes []Node) []Link {
	var links []Link
	var linksDirty []Link
	for _, node := range nodes {
		resp, _ := promQuery(`consulCatalogServiceNodeHealthy{instance="` + node.Instance + `"}`)
		checkPromResponse(resp)
		for _, v := range resp.Data.Result {
			fnn := v.Metric.Node + ":9000"
			value, err := strconv.Atoi(v.Value[1].(string))
			checkerr(err)
			print(v.Metric)
			//linksDirty = append(linksDirty, Link{Source: v.Metric.Instance, Target: fnn, Value: value})
		}
	}
	links = GetUniqueLinks(linksDirty)
	return links
}

// GetUniqueLinks take a array/slice of Link and and returns one with only unique entries
func GetUniqueLinks(linksDirty []Link) []Link {
	var links []Link
	if linksDirty == nil {
		return []Link{}
	}
	if len(linksDirty) <= 1 {
		return linksDirty
	}
	ld, linksDirty := linksDirty[0], linksDirty[1:]
	links = append(links, ld)
	for _, linkDirty := range linksDirty {
		flag := true
		for i, link := range links {
			a := (linkDirty.Source == link.Source)
			b := (linkDirty.Target == link.Target)
			c := (linkDirty.Target == link.Source)
			d := (linkDirty.Source == link.Target)
			if (a && b) || (c && d) {
				flag = false
				links[i].Value++
			}
			if linkDirty.Source == linkDirty.Target {
				flag = false
			}
		}
		if flag {
			links = append(links, linkDirty)
		}
	}
	return links
}

func up(w http.ResponseWriter, r *http.Request) {

	nodesStatus, linksStatus := upFunc()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Status{Nodes: nodesStatus, Links: linksStatus})
}

func consulRaftPeers(w http.ResponseWriter, r *http.Request) {
	resp, _ := promQuery(`consulRaftPeers{job="consul_wolke"}`)
	dudu := resp.Data
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dudu)
}

func consulCatalogServiceNodeHealthy(w http.ResponseWriter, r *http.Request) {
	resp, _ := promQuery(`consulCatalogServiceNodeHealthy`)
	dudu := resp.Data
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dudu)
}

func links(w http.ResponseWriter, r *http.Request) {
	resp, _ := promQuery(`consulCatalogServiceNodeHealthy`)
	dudu := resp.Data
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dudu)
}

// StatusRespWr status struct with embeded http.ResponseWriter
type StatusRespWr struct {
	http.ResponseWriter // We embed http.ResponseWriter
	status              int
}

// WriteHeader writes HTTP header with StatusRespWr
func (w *StatusRespWr) WriteHeader(status int) {
	w.status = status // Store the status for our own use
	w.ResponseWriter.WriteHeader(status)
}

func wrapHandler(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		srw := &StatusRespWr{ResponseWriter: w}
		h.ServeHTTP(srw, r)
		if srw.status >= 400 {
			// 400+ codes are the error codes
			log.Printf("Error status code: %d when serving path: %s",
				srw.status, r.RequestURI)
		}
	}
}

func ls(w http.ResponseWriter, r *http.Request) {
	files, _ := filepath.Glob("*")
	json.NewEncoder(w).Encode(files)
}


func main() {
	flag.Parse()

	fmt.Println("Start panopticon")
	//fs := http.FileServer(http.Dir("static"))
	http.HandleFunc("/", wrapHandler(http.StripPrefix("/", http.FileServer(http.Dir("static")))))
	http.HandleFunc("/ls", ls)
	http.HandleFunc("/api/links", links)
	http.HandleFunc("/api/health", index)
	http.HandleFunc("/api/consul/up", up)
	http.HandleFunc("/api/consul/peers", consulRaftPeers)
	http.HandleFunc("/api/consul/node_healthy", consulCatalogServiceNodeHealthy)
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}
