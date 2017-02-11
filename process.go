package main

import (
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

	if weaveHealth.Health == api.GREEN &&
		glusterHealth.Health == api.GREEN &&
		consulHealth.Health == api.GREEN {
		return api.HS{ClusterState: api.GREEN}, nil
	} else {
		return api.HS{ClusterState: api.RED}, nil
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
	health.Health = 2
	up, err := api.FetchServiceUp(f, api.ConsulUp, promhost)
	if err != nil {
		log.Error(err)
	}
	health.ConsulUp = up.Status

	// get and check consul_raft_peers
	raftPeers, err := api.FetchPromGauge(f, promhost, api.ConsulRaftPeers)
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
				log.Errorf("RaftPeers is %v and expected %v raft peers", peers, api.ClusterNodeCount)
				break
			}
		}
	}
	if peerLen == peerCount && peerLen != 0 && api.ClusterNodeCount == peers {
		health.ConsulRaftPeers = true
	}

	// get and check consul_serf_lan_members
	serfMembers, err := api.FetchPromGauge(f, promhost, api.ConsulSerfLanMembers)
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
			log.Errorf("SerfLanMembers is %v and expected %v members", members, api.ClusterNodeCount)
			break
		}
	}
	if serfLen == serfCount && serfLen != 0 && api.ClusterNodeCount == members {
		health.ConsulSerfMembers = true
	}

	// get and check consul_health_node_status
	healthNodeStatus, err := api.FetchPromGauge(f, promhost, api.ConsulHealthNodeStatus)
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
	raftLeaderResult, err := api.FetchPromGauge(f, promhost, api.ConsulRaftLeader)
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
