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

// Node struct represents json node response
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

type ClusterStatus uint8

const (
	NULL_STATE ClusterStatus = iota
	HEALTHY    ClusterStatus = iota
	WARNING    ClusterStatus = iota
	CRITICAL   ClusterStatus = iota
)

// Consulhealth representates the health state of consul in the  cluster
type ConsulHealth struct {
	Health                 StateType // 0,1,2
	ConsulUp               bool
	ConsulRaftPeers        bool
	ConsulSerfMembers      bool
	ConsulRaftLeader       bool
	ConsulHealthNodeStatus bool
}

// GlusterHealth representates the health state of glusterfs in the  cluster
type GlusterHealth struct {
	Health                     StateType // 0,1,2
	GlusterUp                  bool
	GlusterPeersConnected      bool
	GlusterSuccessfullyMounted bool
	GlusterMountWriteable      bool
}

type WeaveHealth struct {
	Health      StateType // 0,1,2
	Established int64     // number of establised connections should be node count -1
	Connecting  int64
	Failed      int64
	Pending     int64
	Retrying    int64
}

type StateType uint8

const (
	GREEN  StateType = iota
	ORANGE StateType = iota
	RED    StateType = iota
)

type Service struct {
	ServiceState StateType
	Name         string
	Instance     string
}

type HS struct {
	ClusterState   StateType
	Services       []Service
	FailedServices []Service
}
