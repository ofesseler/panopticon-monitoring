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

type ConsulHealth struct {
	Health           int    // 0,1,2
	ConsulUp         bool
	ConsulRaftPeers  bool
	ConsulSerf       bool
	consulRaftLeader bool
}

func FetchConsulHealth(f Fetcher, promhost string) (ConsulHealth, error) {
	var health ConsulHealth
	health.Health = 2
	up, err := FetchServiceUp(f, ConsulUp, promhost)
	if err != nil {
		log.Error(err)
	}

	health.ConsulUp = up.Status

	raftPeers, err := FetchRaftPeers(f, promhost)
	peerLen := len(raftPeers)
	peerCount := 0
	for _,peer := range raftPeers {
		if peer.Value != int64(peerLen) {
			peerCount++
		}
	}
	if peerLen == peerCount {
		health.ConsulRaftPeers = true
	}

	return health, nil
}

type PromQRCount struct {
	Name string
	Job string
	Instance string
	Value int64
}

type ServiceHealth struct {
	Name string // service name
	Instance string //instance name
	Healthy int // 0 (green),1 (orange),2 (red)
}

func FetchRaftPeers(f Fetcher, promHost string) ([]PromQRCount, error) {

	var raftStatusList []PromQRCount

	raftPeerResponse, err := f.PromQuery(ConsulRaftPeers, promHost)
	if err != nil {
		log.Error(err)
		return []PromQRCount{}, err
	}
	for _, result := range raftPeerResponse.Data.Result {
		peers, err := strconv.ParseInt(result.Value[1].(string), 10, 32)
		if err != nil {
			log.Error(err)
			//raftStatus[i].Value = 0
		}
		raftStatusList = append(raftStatusList, PromQRCount{
			Instance: result.Metric.Instance,
			Job: result.Metric.Job,
			Name: result.Metric.Name,
			Value: int64(peers),
		})
	}

	return raftStatusList, nil
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
			return healthSummary, err
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

func mapBoolValueToHeatshStatus(resp StatusCheckReceived) (HealthStatus, error) {
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
			})
		} else {
			failureNodes = append(failureNodes, Health{
				Instance: v.Metric.Instance,
				Query:    v.Metric.Name,
				Ok:       status,
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
	if err != nil && checkPromResponse(resp) {
		return healthStatus, err
	}

	healthStatus, err = mapBoolValueToHeatshStatus(resp)
	if err != nil {
		log.Error(err)
		return healthStatus, err
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
