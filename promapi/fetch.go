package promapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	log "github.com/Sirupsen/logrus"
)

const (
	// Up defines Prometheus metric for targets
	Up = "up"
	// ConsulUp defines respective Prometheus query string
	ConsulUp = "consul_up"
	// ConsulRaftPeers defines respective Prometheus query string
	ConsulRaftPeers = "consul_raft_peers"
	// GlusterUp defines respective Prometheus query string
	GlusterUp = "gluster_up"
	// GlusterPeersConnected defines respective Prometheus query string
	GlusterPeersConnected = "gluster_peers_connected"
	// GlusterHealInfoFilesCount defines respective Prometheus query string
	GlusterHealInfoFilesCount = "gluster_heal_info_files_count"
	// GlusterVolumeWriteable defines respective Prometheus query string
	GlusterVolumeWriteable = "gluster_volume_writeable"
	// GlusterMountSuccessful defines respective Prometheus query string
	GlusterMountSuccessful = "gluster_mount_successful"
	// NodeSupervisorUp defines respective Prometheus query string
	NodeSupervisorUp = "node_supervisor_up"
	// ConsulHealthNodeStatus defines respective Prometheus query string
	ConsulHealthNodeStatus = "consul_health_node_status"
	// ConsulRaftLeader defines respective Prometheus query string
	ConsulRaftLeader = "consul_raft_leader"
	// ConsulSerfLanMembers defines respective Prometheus query string
	ConsulSerfLanMembers = "consul_serf_lan_members"
	// WeaveConnections defines respective Prometheus query string
	WeaveConnections = "weave_connections"
	// NodeLoad15 defines 15m load average
	NodeLoad15 = "node_load15"
	// NodeMemoryMemTotal Memory information field MemTotal.
	NodeMemoryMemTotal = "node_memory_MemTotal"
	// NodeMemoryAvailible Memory information field MemAvailable.
	NodeMemoryAvailible = "node_memory_MemAvailable"
	// MachineCPUCores Number of CPU cores on the machine.
	MachineCPUCores = "machine_cpu_cores"
)

var (
	// ClusterNodeCount represents count of Nodes in Cluster. Set py CMD-parameter at start.
	ClusterNodeCount int
	fatalMetrics     = []string{ConsulUp, GlusterUp}
	warningMetrics   = []string{Up, NodeSupervisorUp}
)

// FetchWeaveConnectionGauges fetches connection values from prometheus API, given the PrometheusFetcher is used
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

// FetchPromInt64 used to fetch int64 values from Prometheus with PrometheusFetcher
func FetchPromInt64(f Fetcher, promHost string, metric string) ([]PromQR, error) {

	var resultMetricList []PromQR

	promResponse, err := f.PromQuery(metric, promHost)
	if err != nil {
		log.Error(err)
		return []PromQR{}, err
	}
	for _, result := range promResponse.Data.Result {
		peers, err := strconv.ParseInt(result.Value[1].(string), 10, 64)
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

// FetchPromFloat64 used to fetch float64 values from Prometheus with PrometheusFetcher
func FetchPromFloat64(f Fetcher, promHost string, metric string) ([]PromQRFloat64, error) {

	var resultMetricList []PromQRFloat64

	promResponse, err := f.PromQuery(metric, promHost)
	if err != nil {
		log.Error(err)
		return []PromQRFloat64{}, err
	}
	for _, result := range promResponse.Data.Result {
		peers, err := strconv.ParseFloat(result.Value[1].(string), 64)
		if err != nil {
			log.Error(err)
			//raftStatus[i].Value = 0
		}
		resultMetricList = append(resultMetricList, PromQRFloat64{
			Node:     result.Metric.Node,
			Job:      result.Metric.Job,
			Name:     result.Metric.Name,
			Value:    peers,
			Instance: result.Metric.Instance,
		})
	}

	return resultMetricList, nil
}

// FetchHealthSummary used to fetch bool metrics (up, consul_up, gluster_up, node_supervisor_up from Prometheus with PrometheusFetcher
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

// FetchServiceUp fetches metric from Promethues with PrometheusFetcher and definded check, such as Up
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

// Fetcher is interface for queries to Prometheus, mainly implemented for testing
type Fetcher interface {
	PromQuery(query string, host string) (StatusCheckReceived, error)
}

// PrometheusFetcher is implementation type for Fetcher
type PrometheusFetcher struct {
}

// PromQuery implementation with PrometheusFetcher to issue queries to Prometheus HTTP API
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
