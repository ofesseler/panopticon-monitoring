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
consul_raft_leader 1

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
# HELP gluster_mount_successful Checks if mountpoint exists, returns 0 or 1
# TYPE gluster_mount_successful gauge
gluster_mount_successful{mountpoint="/mnt/data",volume="localhost:data"} 1
# HELP gluster_volume_writeable Writes and deletes file in Volume and checks if it is writeable
# TYPE gluster_volume_writeable gauge
gluster_volume_writeable{mountpoint="/mnt/data",volume="localhost:data"} 1
# HELP gluster_heal_info_files_count File count of files out of sync, when calling "gluster v heal VOLNAME info".
# TYPE gluster_heal_info_files_count gauge
gluster_heal_info_files_count{volume="localhost:data"} 0
```

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

## Node
```
# HELP node_load15 15m load average.
# TYPE node_load15 gauge
node_load15 0.25
# HELP node_memory_MemTotal Memory information field MemTotal.
# TYPE node_memory_MemTotal gauge
node_memory_MemTotal 2.080411648e+09
# HELP node_memory_MemFree Memory information field MemFree.
# TYPE node_memory_MemFree gauge
node_memory_MemFree 1.19906304e+08
# HELP node_memory_MemAvailable Memory information field MemAvailable.
# TYPE node_memory_MemAvailable gauge
node_memory_MemAvailable 1.38940416e+09
```

## cadvisor
```
# HELP machine_cpu_cores Number of CPU cores on the machine.
# TYPE machine_cpu_cores gauge
machine_cpu_cores 2
```
