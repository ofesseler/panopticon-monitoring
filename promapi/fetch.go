package promapi

import (
	"encoding/json"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const (
	ClusterNodeCount       = 2
	Up                     = "up"
	ConsulUp               = "consul_up"
	ConsulRaftPeers        = "consul_raft_peers"
	GlusterUp              = "gluster_up"
	GlusterPeersConnected  = "gluster_peers_connected"
	NodeSupervisorUp       = "node_supervisor_up"
	ConsulHealthNodeStatus = "consul_health_node_status"
	ConsulRaftLeader       = "consul_raft_leader"
	ConsulSerfLanMembers   = "consul_serf_lan_members"
	WeaveConnections       = "weave_connections"
)

var (
	fatalMetrics   = []string{ConsulUp, GlusterUp}
	warningMetrics = []string{Up, NodeSupervisorUp}
)

func FetchWeaveConnectionGauges(f Fetcher, promHost string, metric string) ([]PromQRWeave, error) {

	var resultMetricList []PromQRWeave

	promResponse, err := f.PromQuery(metric, promHost)
	if err != nil {
		log.Error(err)
		return []PromQRWeave{}, err
	}
	for _, result := range promResponse.Data.Result {
		peers, err := strconv.ParseInt(result.Value[1].(string), 10, 32)
		if err != nil {
			log.Error(err)
			//raftStatus[i].Value = 0
		}
		pw := PromQRWeave{State: result.Metric.State}
		pw.Instance = result.Metric.Instance
		pw.Node = result.Metric.Node
		pw.Job = result.Metric.Job
		pw.Name = result.Metric.Name
		pw.Value = int64(peers)
		resultMetricList = append(resultMetricList, pw)
	}

	return resultMetricList, nil
}

func FetchPromGauge(f Fetcher, promHost string, metric string) ([]PromQR, error) {

	var resultMetricList []PromQR

	promResponse, err := f.PromQuery(metric, promHost)
	if err != nil {
		log.Error(err)
		return []PromQR{}, err
	}
	for _, result := range promResponse.Data.Result {
		peers, err := strconv.ParseInt(result.Value[1].(string), 10, 32)
		if err != nil {
			log.Error(err)
			//raftStatus[i].Value = 0
		}
		resultMetricList = append(resultMetricList, PromQR{
			Node:     result.Metric.Node,
			Job:      result.Metric.Job,
			Name:     result.Metric.Name,
			Value:    int64(peers),
			Instance: result.Metric.Instance,
		})
	}

	return resultMetricList, nil
}

func FetchHealthSummary(f Fetcher, promHost string) (HealthSummary, error) {
	var healthSummary HealthSummary
	upList := []string{Up, ConsulUp, GlusterUp, NodeSupervisorUp}
	count := 0
	for _, v := range upList {
		check, err := FetchServiceUp(f, v, promHost)
		if err != nil {
			log.Error(err)
			return HealthSummary{}, err
		}
		if !check.Status {
			healthSummary.Failed = append(healthSummary.Failed, v)
			log.Warn(v)
		}
		count++
	}
	healthSummary.Status = len(healthSummary.Failed) <= 0
	healthSummary.Services = upList
	return healthSummary, nil
}

func mapBoolValueToHeathStatus(resp StatusCheckReceived) (HealthStatus, error) {
	var (
		healthyNodes []PromQueryRequest
		failureNodes []PromQueryRequest
		healthStatus HealthStatus
	)
	for _, v := range resp.Data.Result {
		status, err := strconv.ParseBool(v.Value[1].(string))
		if err != nil {
			log.Error(err)
			return healthStatus, err
		}
		if status {
			healthyNodes = append(healthyNodes, PromQueryRequest{
				Instance: v.Metric.Instance,
				Query:    v.Metric.Name,
				Ok:       status,
				Job:      v.Metric.Job,
			})
		} else {
			failureNodes = append(failureNodes, PromQueryRequest{
				Instance: v.Metric.Instance,
				Query:    v.Metric.Name,
				Ok:       status,
				Job:      v.Metric.Job,
			})
		}

	}
	healthStatus.FailureNodes = failureNodes
	healthStatus.HealthyNodes = healthyNodes
	healthStatus.Status = (len(failureNodes) == 0) && (len(failureNodes) > -1)
	return healthStatus, nil
}

func FetchServiceUp(f Fetcher, check, promHost string) (HealthStatus, error) {
	var healthStatus HealthStatus
	resp, err := f.PromQuery(check, promHost)
	if err != nil || !checkPromResponse(resp) {
		return HealthStatus{}, err
	}

	healthStatus, err = mapBoolValueToHeathStatus(resp)
	if err != nil {
		log.Error(err)
		return HealthStatus{}, err
	}

	return healthStatus, nil
}

type Fetcher interface {
	PromQuery(query string, host string) (StatusCheckReceived, error)
}

type PrometheusFetcher struct {
}

func (PrometheusFetcher) PromQuery(query, promHost string) (StatusCheckReceived, error) {
	var errorStatus error
	apiURL := fmt.Sprintf("http://%v/api/v1/query", promHost)
	urlValues := url.Values{}
	urlValues.Set("query", query)
	concatedURL := apiURL + "?" + urlValues.Encode()
	checkURL, err := url.Parse(concatedURL)
	if err != nil {
		log.Error(err)
		return StatusCheckReceived{}, err
	}
	request, err := http.NewRequest("GET", checkURL.String(), nil)
	if err != nil {
		log.Error(err)
		return StatusCheckReceived{}, err
	}
	client := &http.Client{
		Timeout: 3 * time.Second,
	}
	resp, err := client.Do(request)
	if err != nil {
		log.Error(err)
		return StatusCheckReceived{}, err
	}
	defer resp.Body.Close()
	body, _ := decodeResponse(resp)

	return body, errorStatus
}

func decodeResponse(response *http.Response) (StatusCheckReceived, ErrorStatus) {
	var errorStatus ErrorStatus
	decoder := json.NewDecoder(response.Body)
	var body StatusCheckReceived
	err := decoder.Decode(&body)
	if body.Status != "success" && err != nil {
		log.Error("Prometheus responded with error:", body.Status)
	}
	return body, errorStatus
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
