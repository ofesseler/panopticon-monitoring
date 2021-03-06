package promapi

// StatusCheckReceived struct represents json response
type StatusCheckReceived struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string   `json:"resultType"`
		Result     []Result `json:"result"`
	} `json:"data"`
}

type Result struct {
	Metric struct {
		Name     string `json:"__name__"`
		Check    string `json:"check"`
		Instance string `json:"instance"`
		Job      string `json:"job"`
		Node     string `json:"node"`
		State    string `json:"state"`
	} `json:"metric"`
	Value []interface{} `json:"value"`
}

// ErrorStatus struct represents json error response
type ErrorStatus struct {
	Status    string `json:"status"`
	ErrorType string `json:"errorType"`
	Error     string `json:"error"`
}

type PromQueryRequest struct {
	Instance string `json:"instance"`
	Job      string `json:"job"`
	Ok       bool   `json:"ok"`
	Query    string `json:"query"`
}

// Node struct represents json node1 response
type Node struct {
	Instance string `json:"instance"` // Instance URL
	Group    int    `json:"group"`    // Group should be 0,1,2 -> green, orange, red
	ID       int    `json:"id"`       // ID ??
}

// Link struct represents json
type Link struct {
	Source string `json:"source"`
	Target string `json:"target"`
	Value  int    `json:"value"`
}

type HealthStatus struct {
	Status       bool               `json:"status"` // Status 0,1,2 maps to health status green (0), orange(1), red (2)
	HealthyNodes []PromQueryRequest `json:"healthyNodes"`
	FailureNodes []PromQueryRequest `json:"failureNodes"`
}

type HealthSummary struct {
	Status   bool
	Services []string
	Failed   []string
}

type PromQR struct {
	Name     string
	Job      string
	Node     string
	Value    int64
	Instance string
}

type PromQRWeave struct {
	PromQR
	State string
}

type PromQRFloat64 struct {
	Name     string
	Job      string
	Node     string
	Instance string
	Value    float64
}

type ClusterStatus uint8

const (
	NULL_STATE ClusterStatus = iota
	HEALTHY    ClusterStatus = iota
	WARNING    ClusterStatus = iota
	CRITICAL   ClusterStatus = iota
)

func GetClusterStatusString(cs ClusterStatus) string {
	var p string
	switch cs {
	case HEALTHY:
		p = "healthy"
	case WARNING:
		p = "warning"
	case CRITICAL:
		p = "critical"
	case NULL_STATE:
		p = "null_state"
	}
	return p
}

// Consulhealth representates the health state of consul in the  cluster
type ConsulHealth struct {
	Health                 ClusterStatus
	ConsulUp               ClusterStatus
	ConsulRaftPeers        ClusterStatus
	ConsulSerfLanMembers   ClusterStatus
	ConsulRaftLeader       ClusterStatus
	ConsulHealthNodeStatus ClusterStatus
}

// GlusterHealth representates the health state of glusterfs in the  cluster
type GlusterHealth struct {
	Health                    ClusterStatus
	GlusterUp                 ClusterStatus
	GlusterPeersConnected     ClusterStatus
	GlusterVolumeWriteable    ClusterStatus
	GlusterMountSuccessful    ClusterStatus
	GlusterHealInfoFilesCount ClusterStatus
}

type WeaveHealth struct {
	Health      ClusterStatus
	Established int64 // number of establised connections should be node1 count -1
	Connecting  int64
	Failed      int64
	Pending     int64
	Retrying    int64
}

type HostHealth struct {
	Health     ClusterStatus
	Load15     ClusterStatus
	MemoryFree ClusterStatus
}

type Service struct {
	ServiceState ClusterStatus
	Name         string
	Instance     string
}

type HS struct {
	ClusterState   ClusterStatus
	Services       []Service
	FailedServices []Service
}
