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
	Up               = "up"
	ConsulUp         = "consul_up"
	ConsulRaftPeers  = "consul_raft_peers"
	GlusterUp        = "gluster_up"
	NodeSupervisorUp = "node_supervisor_up"
)

var (
	fatalMetrics     = []string{ConsulUp, GlusterUp}
	warningMetrics   = []string{Up, NodeSupervisorUp}
)

type ConsulHealth struct {
	Health int // 0,1,2
	ConsulUp bool
	ConsulRaftPeers bool
	ConsulSerf bool
	consulRaftLeader bool
}

func CheckConsulHealth(promhost string) {
	var health ConsulHealth
	up, err := CheckUp(*promhost, ConsulUp)
	if err != nil {
		log.Error(err)
	}
	health.ConsulUp = up.Status

	raftPeersStatus, err := promQuery(ConsulRaftPeers, promhost)


	for _, v := range raftPeersStatus.Data.Result{
		peers, err := strconv.ParseInt(v.Value[1].(string),10,64)
		if err != nil {
			log.Error(err)
		}
		if peers != 4 {
			log.Errorf("Peers should be 4 and is: %v", peers)
		}
		
	}

}

func FetchHealthSummary(promHost string) (HealthSummary, error) {
	var healthSummary HealthSummary
	upList := []string{Up, ConsulUp, GlusterUp, NodeSupervisorUp}
	count := 0
	for _, v := range upList {
		check, err := CheckUp(promHost, v)
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

func mapResponseToHealthStatus(resp StatusCheckReceived) (HealthStatus, error) {
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

func CheckUp(promHost, check string) (HealthStatus, error) {
	var healthStatus HealthStatus

	resp, err := promQuery(check, promHost)
	if err != nil && checkPromResponse(resp) {
		return healthStatus, err
	}

	healthStatus, err = mapResponseToHealthStatus(resp)
	if err != nil {
		log.Error(err)
		return healthStatus, err
	}

	return healthStatus, nil
}

func promQuery(query, promHost string) (StatusCheckReceived, error) {
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

