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

func computeUp(f api.Fetcher, check string, promHost string) (api.ClusterStatus, error) {
	up, err := api.FetchServiceUp(f, check, promHost)
	if err != nil {
		log.Error(err)
		return api.NULL_STATE, err
	}
	qr := QuorumRate{}
	upRated, err := qr.Rater(len(up.HealthyNodes), api.ClusterNodeCount)
	if err != nil {
		log.Error(err)
	}
	return upRated, nil
}

func ProcessGlusterHealthSummary(f api.Fetcher, promhost string) (api.GlusterHealth, error) {
	gh := api.GlusterHealth{Health: 2}

	// GlusterUP test
	var err error
	gh.GlusterUp, err = computeUp(f, api.GlusterUp, promhost)
	if err != nil {
		log.Error(err)
	}

	// Peers connected test
	peersConnectedList, err := api.FetchPromGauge(f, promhost, api.GlusterPeersConnected)
	if err != nil {
		log.Error(err)
	}
	peerRate := GlusterPeerRate{}
	gh.GlusterPeersConnected = computeCountersFromPromQRs(peerRate, api.ClusterNodeCount, peersConnectedList)

	healFilesStatus := api.HEALTHY
	healFilesCount, err := api.FetchPromGauge(f, promhost, api.GlusterHealInfoFilesCount)
	if err != nil {
		log.Error(err)
	}
	for _, file := range healFilesCount {
		if file.Value != 0 {
			healFilesStatus = api.CRITICAL
			break
		}
	}
	gh.GlusterHealInfoFilesCount = healFilesStatus

	mountSuccessful, err := api.FetchServiceUp(f, api.GlusterMountSuccessful, promhost)
	if err != nil {
		log.Error(err)
	}
	gh.GlusterMountSuccessful = rateBool(mountSuccessful.Status, api.CRITICAL)

	volumeWriteable, err := api.FetchServiceUp(f, api.GlusterVolumeWriteable, promhost)
	if err != nil {
		log.Error(err)
	}
	gh.GlusterVolumeWriteable = rateBool(volumeWriteable.Status, api.CRITICAL)

	gh.Health = computeHealthStatus(gh.GlusterUp, gh.GlusterPeersConnected, gh.GlusterHealInfoFilesCount, gh.GlusterHealInfoFilesCount, gh.GlusterMountSuccessful, gh.GlusterVolumeWriteable)
	return gh, nil
}

func ProcessConsulHealthSummary(f api.Fetcher, promHost string) (api.ConsulHealth, error) {
	// check Consul reachable and running
	var (
		health api.ConsulHealth
		err    error
		qr     = QuorumRate{}
		pr     = PromRate{}
	)

	health.ConsulUp, err = computeUp(f, api.ConsulUp, promHost)
	if err != nil {
		log.Error(err)
	}

	// get and check consul_raft_peers
	raftPeers, err := api.FetchPromGauge(f, promHost, api.ConsulRaftPeers)
	if err != nil {
		log.Error(err)
	}
	health.ConsulRaftPeers = computeCountersFromPromQRs(pr, api.ClusterNodeCount, raftPeers)

	// get and check consul_serf_lan_members
	serfMembers, err := api.FetchPromGauge(f, promHost, api.ConsulSerfLanMembers)
	if err != nil {
		log.Error(err)
	}
	health.ConsulSerfLanMembers = computeCountersFromPromQRs(pr, api.ClusterNodeCount, serfMembers)

	// get and check consul_health_node_status
	healthNodeStatus, err := api.FetchPromGauge(f, promHost, api.ConsulHealthNodeStatus)
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
	raftLeaderResult, err := api.FetchPromGauge(f, promHost, api.ConsulRaftLeader)
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
	reference := ireference.(int)
	q := QuorumRate{}
	return q.Rater(int(value.Value), reference)
}

type GlusterPeerRate struct{}

func (r GlusterPeerRate) Rater(ivalue, ireference interface{}) (api.ClusterStatus, error) {
	prom := ivalue.(api.PromQR)
	reference := ireference.(int)
	value := int(prom.Value)
	var cs api.ClusterStatus
	var err error = nil

	// adds +1 to value to calculate quorum
	value += 1
	if value == reference {
		cs = api.HEALTHY
	} else if value >= (reference/2)+1 {
		cs = api.WARNING
	} else if value < (reference/2)+1 {
		cs = api.CRITICAL
	}
	return cs, err
}

func computeCountersFromPromQRs(r Rate, reference int, promQRs []api.PromQR) api.ClusterStatus {
	var (
		hCounter int = 0
		wCounter int = 0
		cCounter int = 0
		status   api.ClusterStatus
	)

	for _, promQR := range promQRs {
		computedStatus, err := r.Rater(promQR, reference)
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
	if hCounter == reference {
		status = api.HEALTHY
	} else if wCounter >= (reference/2)+1 || hCounter >= (reference/2)+1 {
		status = api.WARNING
	} else if cCounter >= (reference / 2) {
		status = api.CRITICAL
	}
	return status
}
