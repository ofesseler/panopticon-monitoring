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
	ClusterNodeCount       = 4
	Up                     = "up"
	ConsulUp               = "consul_up"
	ConsulRaftPeers        = "consul_raft_peers"
	GlusterUp              = "gluster_up"
	NodeSupervisorUp       = "node_supervisor_up"
	ConsulHealthNodeStatus = "consul_health_node_status"
	ConsulRaftLeader       = "consul_raft_leader"
	ConsulSerfLanMembers   = "consul_serf_lan_members"
)

var (
	fatalMetrics   = []string{ConsulUp, GlusterUp}
	warningMetrics = []string{Up, NodeSupervisorUp}
)


func ProcessConsulHealthSummary(f Fetcher, promhost string) (ConsulHealth, error) {
	// check Consul reachable and running
	var health ConsulHealth
	health.Health = 2
	up, err := FetchServiceUp(f, ConsulUp, promhost)
	if err != nil {
		log.Error(err)
	}

	health.ConsulUp = up.Status

	// get and check consul_raft_peers
	raftPeers, err := FetchPromGauge(f, promhost, ConsulRaftPeers)
	peerLen := len(raftPeers)
	peerCount := 0
	var peers int64 = -1
	for _, peer := range raftPeers {
		if peer.Value != int64(peerLen) {
			peerCount++
			if peers == -1 {
				peers = peer.Value
			}
			if peers != peer.Value {
				log.Errorf("RaftPeers is %v and expected %v raft peers", peers, ClusterNodeCount)
				break
			}
		}
	}
	if peerLen == peerCount && peerLen != 0 && ClusterNodeCount == peers {
		health.ConsulRaftPeers = true
	}

	// get and check consul_serf_lan_members
	serfMembers, err := FetchPromGauge(f, promhost, ConsulSerfLanMembers)
	if err != nil {
		log.Error(err)
	}
	serfLen := len(serfMembers)
	serfCount := 0
	var members int64 = -1
	for _, member := range raftPeers {
		if member.Value != int64(peerLen) {
			serfCount++
		}
		if members == -1 {
			members = member.Value
		}
		if members != member.Value {
			log.Errorf("SerfLanMembers is %v and expected %v members", members, ClusterNodeCount)
			break
		}
	}
	if serfLen == serfCount && serfLen != 0 && ClusterNodeCount == members {
		health.ConsulSerfMembers = true
	}

	// get and check consul_health_node_status
	healthNodeStatus, err := FetchPromGauge(f, promhost, ConsulHealthNodeStatus)
	if err != nil {
		log.Error(err)
	}
	var nodeStatusCount int = 0
	for _, node := range healthNodeStatus {
		if node.Value == 1 {
			nodeStatusCount++
		} else {
			log.Infof("SerfLanMembers is not healthy: %v", node)
		}
	}
	if nodeStatusCount == len(healthNodeStatus) {
		health.ConsulHealthNodeStatus = true
	}

	// Is there a leader?
	raftLeaderResult, err := FetchPromGauge(f, promhost, ConsulRaftLeader)
	if err != nil {
		log.Error(err)
	}
	var instancesWithLeader int = 0
	for _, node := range raftLeaderResult {
		if node.Value == 1 {
			instancesWithLeader++
		} else {
			log.Infof("SerfLanMembers is not healthy: %v", node)
		}
	}
	if instancesWithLeader == len(raftLeaderResult) {
		health.ConsulRaftLeader = true
	}

	if health.ConsulUp && health.ConsulRaftPeers && health.ConsulSerfMembers && health.ConsulHealthNodeStatus && health.ConsulRaftLeader {
		health.Health = 0
	}

	return health, nil
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
			Node:  result.Metric.Node,
			Job:   result.Metric.Job,
			Name:  result.Metric.Name,
			Value: int64(peers),
		})
	}

	return resultMetricList, nil
}

func FetchHealthSummary(promHost string) (HealthSummary, error) {
	var healthSummary HealthSummary
	upList := []string{Up, ConsulUp, GlusterUp, NodeSupervisorUp}
	count := 0
	for _, v := range upList {
		var f Fetcher = PrometheusFetcher{}
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
		healthyNodes []Health
		failureNodes []Health
		healthStatus HealthStatus
	)
	for _, v := range resp.Data.Result {
		status, err := strconv.ParseBool(v.Value[1].(string))
		if err != nil {
			log.Error(err)
			return healthStatus, err
		}
		if status {
			healthyNodes = append(healthyNodes, Health{
				Instance: v.Metric.Instance,
				Query:    v.Metric.Name,
				Ok:       status,
				Job:      v.Metric.Job,
			})
		} else {
			failureNodes = append(failureNodes, Health{
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

type PrometheusFetcher struct{}

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
