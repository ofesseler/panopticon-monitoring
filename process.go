package main

import (
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	api "github.com/ofesseler/panopticon/promapi"
)

func ProcessHealthSummary(f api.Fetcher, promHost string) (api.HS, error) {

	glusterHealth, err := ProcessGlusterHealthSummary(f, promHost)
	if err != nil {
		log.Error(err)
	}

	consulHealth, err := ProcessConsulHealthSummary(f, promHost)
	if err != nil {
		log.Error(err)
	}

	weaveHealth, err := ProcessWeaveHealthSummary(f, promHost)
	if err != nil {
		log.Error(err)
	}

	if weaveHealth.Health == api.HEALTHY &&
		glusterHealth.Health == api.HEALTHY &&
		consulHealth.Health == api.HEALTHY {
		return api.HS{ClusterState: api.HEALTHY}, nil
	} else {
		return api.HS{ClusterState: api.HEALTHY}, nil
	}

	return api.HS{}, nil
}

func ProcessWeaveHealthSummary(f api.Fetcher, promhost string) (api.WeaveHealth, error) {
	wh := api.WeaveHealth{Health: 2}
	weaveTest := true

	// weave connections
	connList, err := api.FetchWeaveConnectionGauges(f, promhost, api.WeaveConnections)
	if err != nil {
		log.Error(err)
		weaveTest = false
	}

	for _, con := range connList {
		switch con.State {
		case "connecting":
			wh.Connecting += con.Value
		case "established":
			wh.Established += con.Value
		case "pending":
			wh.Pending += con.Value
		case "failed":
			wh.Failed += con.Value
		case "retrying":
			wh.Retrying += con.Value
		}
	}

	if wh.Pending > 0 || wh.Retrying > 0 || wh.Failed > 0 {
		weaveTest = false
	}
	if weaveTest {
		wh.Health = 0
	}
	return wh, nil
}

func ProcessGlusterHealthSummary(f api.Fetcher, promhost string) (api.GlusterHealth, error) {
	gh := api.GlusterHealth{Health: 2}
	glusterTest := true

	// GlusterUP test
	up, err := api.FetchPromGauge(f, promhost, api.GlusterUp)
	if err != nil {
		log.Error(err)
		glusterTest = false
	}
	for _, peer := range up {
		if peer.Value != 1 {
			glusterTest = false
		}
	}
	gh.GlusterUp = glusterTest

	// Peers connected test
	peersConnectedList, err := api.FetchPromGauge(f, promhost, api.GlusterPeersConnected)
	if err != nil {
		log.Error(err)
	}
	// reset glusterTest
	glusterTest = true
	peersLen := len(peersConnectedList)
	if peersLen != api.ClusterNodeCount {
		glusterTest = false
		log.Errorf("Not all Cluster nodes are reachable expected %v and reached %v", api.ClusterNodeCount, peersLen)
	}

	for _, peer := range peersConnectedList {
		if int64(peersLen-1) != peer.Value {
			log.Errorf("cluster_peers_connected value %v and scaped peers( %v ) don't match", peer.Value, peersLen-1)
			glusterTest = false
		}
	}
	if glusterTest {
		gh.GlusterPeersConnected = true
	}

	// TODO
	//gh.GlusterMountWriteable = true
	//gh.GlusterSuccessfullyMounted = true

	//if gh.GlusterUp && gh.GlusterPeersConnected && gh.GlusterSuccessfullyMounted && gh.GlusterMountWriteable {
	//	gh.Health = 0
	//}

	if gh.GlusterUp && gh.GlusterPeersConnected {
		gh.Health = 0
	}
	return gh, nil
}

func ProcessConsulHealthSummary(f api.Fetcher, promhost string) (api.ConsulHealth, error) {
	// check Consul reachable and running
	var health api.ConsulHealth
	//health.Health = 2

	up, err := api.FetchServiceUp(f, api.ConsulUp, promhost)
	if err != nil {
		log.Error(err)
	}
	qr := QuorumRate{}
	health.ConsulUp, err = qr.Rater(len(up.HealthyNodes), api.ClusterNodeCount)
	if err != nil {
		log.Error(err)
	}
	// get and check consul_raft_peers
	raftPeers, err := api.FetchPromGauge(f, promhost, api.ConsulRaftPeers)
	if err != nil {
		log.Error(err)
	}

	health.ConsulRaftPeers = computeCountersFromPromQRs(raftPeers)

	// get and check consul_serf_lan_members
	serfMembers, err := api.FetchPromGauge(f, promhost, api.ConsulSerfLanMembers)
	if err != nil {
		log.Error(err)
	}

	health.ConsulSerfLanMembers = computeCountersFromPromQRs(serfMembers)

	// get and check consul_health_node_status
	healthNodeStatus, err := api.FetchPromGauge(f, promhost, api.ConsulHealthNodeStatus)
	if err != nil {
		log.Error(err)
	}

	var nodeStatusCount int = 0
	var countByNode = make(map[string]int)

	for _, node := range healthNodeStatus {
		countByNode[node.Node] += 1

		if node.Value == 1 {
			nodeStatusCount++
		} else {
			log.Errorf("health node status  is not healthy: %v", node)
		}
	}
	var nodeCounts []int
	nodeCounts = append(nodeCounts, len(countByNode))
	for _, v := range countByNode {
		nodeCounts = append(nodeCounts, v)
	}

	health.ConsulHealthNodeStatus = computeMetricStatus(qr, nodeCounts...)

	// Is there a leader?
	raftLeaderResult, err := api.FetchPromGauge(f, promhost, api.ConsulRaftLeader)
	if err != nil {
		log.Error(err)
	}
	var instancesWithLeader int = 0
	for _, node := range raftLeaderResult {
		if node.Value == 1 {
			instancesWithLeader++
		} else {
			log.Infof("Raft Leader is not healthy: %v", node)
		}
	}
	health.ConsulRaftLeader = computeMetricStatus(qr, instancesWithLeader, len(raftLeaderResult))

	health.Health = computeHealthStatus(health.ConsulUp, health.ConsulRaftPeers, health.ConsulHealthNodeStatus, health.ConsulRaftLeader, health.ConsulSerfLanMembers)

	return health, nil
}

type Rate interface {
	Rater(value, reference interface{}) (api.ClusterStatus, error)
}

type QuorumRate struct{}

func (r QuorumRate) Rater(ivalue, ireference interface{}) (api.ClusterStatus, error) {
	value, ok := ivalue.(int)
	if !ok {
		return api.NULL_STATE, errors.New(fmt.Sprintf("Expected int at parameter ivalue and got: %v", ivalue))
	}
	reference, ok := ireference.(int)
	if !ok {
		return api.NULL_STATE, errors.New(fmt.Sprintf("Expected int at parameter ireference and got:%v", ireference))
	}
	var cs api.ClusterStatus
	var err error = nil
	if value == reference {
		cs = api.HEALTHY
	} else if value >= (reference/2)+1 {
		cs = api.WARNING
	} else if value < (reference/2)+1 {
		cs = api.CRITICAL
	}
	return cs, err
}

func rateBool(value bool, falseState api.ClusterStatus) api.ClusterStatus {
	if value {
		return api.HEALTHY
	}
	return falseState
}

func computeMetricStatus(r Rate, counter ...int) api.ClusterStatus {
	metricStatus := api.HEALTHY
	for _, c := range counter {
		rating, err := r.Rater(c, api.ClusterNodeCount)
		if err != nil {
			log.Error(err)
		}
		if metricStatus < rating {
			metricStatus = rating
		}
	}
	return metricStatus
}

func computeHealthStatus(statuses ...api.ClusterStatus) api.ClusterStatus {
	healthStatus := api.HEALTHY
	for _, status := range statuses {
		if healthStatus < status {
			healthStatus = status
		}
	}
	return healthStatus
}

type PromRate struct{}

func (r PromRate) Rater(ivalue, ireference interface{}) (api.ClusterStatus, error) {
	value := ivalue.(api.PromQR)
	q := QuorumRate{}
	return q.Rater(int(value.Value), api.ClusterNodeCount)
}

func computeCountersFromPromQRs(promQRs []api.PromQR) api.ClusterStatus {
	var (
		hCounter  int = 0
		wCounter  int = 0
		cCounter  int = 0
		status    api.ClusterStatus
		nodeCount int = api.ClusterNodeCount
	)

	for _, promQR := range promQRs {
		r := PromRate{}
		computedStatus, err := r.Rater(promQR, api.ClusterNodeCount)
		if err != nil {
			log.Error(err)
		}

		switch computedStatus {
		case api.HEALTHY:
			hCounter++
		case api.WARNING:
			wCounter++
		case api.CRITICAL:
			cCounter++
		}
	}
	if hCounter == nodeCount {
		status = api.HEALTHY
	} else if wCounter >= (nodeCount/2)+1 || hCounter >= (nodeCount/2)+1 {
		status = api.WARNING
	} else if cCounter >= (nodeCount / 2) {
		status = api.CRITICAL
	}
	return status
}
