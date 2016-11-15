package main

import (
	"fmt"
	"log"
	"time"
	"net/http"
	"encoding/json"
	"net/url"
	"strconv"
)

type Status_Check_Received struct {
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
			       Value  []interface{} `json:"value"`
		       } `json:"result"`
	       } `json:"data"`
}

type Error_Status struct {
	Status    string `json:"status"`
	ErrorType string `json:"errorType"`
	Error     string `json:"error"`
}

type Node struct {
	Instance string     `json:"instance"`
	Group    int        `json:"group"`
	Id       int            `json:"id"`
}
type Link struct {
	Source string        `json:"source"`
	Target string        `json:"target"`
	Value  int        `json:"value"`
}
type Status struct {
	Nodes []Node        `json:"nodes"`
	Links []Link        `json:"links"`
}

func checkerr(err error) {
	if err != nil {
		fmt.Println("ohh jee")
		log.Fatal(err)
	}
}

func checkPromResponse(resp Status_Check_Received) bool {
	if resp.Status != "success" {
		log.Fatal("prometheus request failed", resp)
		return false
	}
	if resp.Data.ResultType != "vector" {
		log.Fatal("prometheus request returned wrong ResultType (other than 'vector'): ", resp.Data.ResultType)
		return false
	}
	return true
}

func promQuery(query string) (Status_Check_Received, Error_Status) {
	var error_status Error_Status
	api_url := "http://localhost:9090/api/v1/query"
	url_values := url.Values{}
	url_values.Set("query", query)
	check_url, err := url.Parse(api_url + "?" + url_values.Encode())
	fmt.Println(check_url)
	request, err := http.NewRequest("GET", check_url.String(), nil)
	client := &http.Client{
		Timeout: 3 * time.Second,
	}
	resp, err := client.Do(request)

	checkerr(err)
	defer resp.Body.Close()
	body, _ := decodeResponse(resp)

	return body, error_status
}

func decodeResponse(response *http.Response) (Status_Check_Received, Error_Status) {
	var error_status Error_Status
	decoder := json.NewDecoder(response.Body)
	var body Status_Check_Received
	err := decoder.Decode(&body)
	checkerr(err)
	if body.Status != "success" {
		log.Fatal("Prometheus responded with error:", body.Status)
	}
	return body, error_status
}

func index(w http.ResponseWriter, r *http.Request) {
	status := make(map[string]string)
	status["status"] = "ok"
	status["heath"] = "green"
	json.NewEncoder(w).Encode(status)
}

func up_func() ([]Node, []Link) {
	resp, _ := promQuery(`up{job="consul_wolke"}`)
	checkPromResponse(resp)
	var nodes_status []Node
	var links_status []Link
	for _, v := range resp.Data.Result {
		value, _ := strconv.Atoi(v.Value[1].(string))
		node := Node{Instance:v.Metric.Instance, Group:value}
		nodes_status = append(nodes_status, node)
	}
	links_status = getLinksForNodes(nodes_status)
	return nodes_status, links_status
}

func getLinksForNodes(nodes []Node) ([]Link) {
	var links []Link
	var links_dirty []Link
	for _, node := range nodes {
		fmt.Println("instance: ", node.Instance)
		resp, _ := promQuery(`consul_catalog_service_node_healthy{instance="` + node.Instance + `"}`)
		checkPromResponse(resp)
		for _, v := range resp.Data.Result {
			fnn := v.Metric.Node + ":9000"
			value, err := strconv.Atoi(v.Value[1].(string))
			checkerr(err)
			links_dirty = append(links_dirty, Link{Source:v.Metric.Instance, Target:fnn, Value:value})
		}
	}
	links = GetUniqueLinks(links_dirty)
	return links
}

func GetUniqueLinks(links_dirty []Link) ([]Link) {
	var links []Link
	if (links_dirty== nil) {
		return []Link{}
	}
	if (len(links_dirty) <= 1) {
		return links_dirty
	}
	ld, links_dirty := links_dirty[0], links_dirty[1:]
	links = append(links, ld)
	for _, link_dirty := range links_dirty {
		flag := true
		for i, link := range links {
			a := (link_dirty.Source == link.Source)
			b := (link_dirty.Target == link.Target)
			c := (link_dirty.Target == link.Source)
			d := (link_dirty.Source == link.Target)
			if ((a && b) || (c && d)) {
				flag = false
				links[i].Value += 1
			}
			if link_dirty.Source == link_dirty.Target {
				flag = false
			}
		}
		if flag {
			links = append(links, link_dirty)
		}
	}
	return links
}

func up(w http.ResponseWriter, r *http.Request) {

	nodes_status, links_status := up_func()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Status{Nodes:nodes_status, Links:links_status})
}

func consul_raft_peers(w http.ResponseWriter, r *http.Request) {
	resp, _ := promQuery(`consul_raft_peers{job="consul_wolke"}`)
	dudu := resp.Data
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dudu)
}

func consul_catalog_service_node_healthy(w http.ResponseWriter, r *http.Request) {
	resp, _ := promQuery(`consul_catalog_service_node_healthy`)
	dudu := resp.Data
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dudu)
}

func links (w http.ResponseWriter, r *http.Request) {
	resp, _ := promQuery(`consul_catalog_service_node_healthy`)
	dudu := resp.Data
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dudu)
}
func main() {
	fmt.Println("Start")
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/", fs)
	http.HandleFunc("/api/links", links)
	http.HandleFunc("/api/health", index)
	http.HandleFunc("/api/consul/up", up)
	http.HandleFunc("/api/consul/peers", consul_raft_peers)
	http.HandleFunc("/api/consul/node_healthy", consul_catalog_service_node_healthy)
	log.Fatal(http.ListenAndServe(":8888", nil))
}
