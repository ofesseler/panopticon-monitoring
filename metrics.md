# Metriken Besprochen

## Consul
```
# HELP consul_health_node_status Status of health checks associated with a node.
# TYPE consul_health_node_status gauge
consul_health_node_status{check="serfHealth",node="wolke1.leomedia.local"} 1
consul_health_node_status{check="serfHealth",node="wolke2.leomedia.local"} 1
consul_health_node_status{check="serfHealth",node="wolke3.leomedia.local"} 1
consul_health_node_status{check="serfHealth",node="wolke4.leomedia.local"} 1
# HELP consul_raft_leader Does Raft cluster have a leader (according to this node).
# TYPE consul_raft_leader gauge
consul_raft_leader{leader="192.168.142.112:8300"} 1
# HELP consul_raft_peers How many peers (servers) are in the Raft cluster.
# TYPE consul_raft_peers gauge
consul_raft_peers 4
# HELP consul_serf_lan_members How many members are in the cluster.
# TYPE consul_serf_lan_members gauge
consul_serf_lan_members 4
# HELP consul_up Was the last query of Consul successful.
# TYPE consul_up gauge
consul_up 1
```

## Gluster
```
# HELP gluster_peers_connected Is peer connected to gluster cluster.
# TYPE gluster_peers_connected gauge
gluster_peers_connected 3
# HELP gluster_up Was the last query of Gluster successful.
# TYPE gluster_up gauge
gluster_up 1
```
+ Mountpoints
+ Mountpoints beschreibbar?

## Weave
```
# HELP weave_connections Number of peer-to-peer connections.
# TYPE weave_connections gauge
weave_connections{state="connecting"} 0
weave_connections{state="established"} 3
weave_connections{state="failed"} 0
weave_connections{state="pending"} 0
weave_connections{state="retrying"} 0
```
failed, pending und retrying sollten immer 0 sein. sonst schelcht.
Connecting kann auch mal anders sein, vorallem ncah error. vielleicht warning?
