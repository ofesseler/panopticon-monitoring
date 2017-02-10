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
	} `json:"metric"`
	Value []interface{} `json:"value"`
}

// ErrorStatus struct represents json error response
type ErrorStatus struct {
	Status    string `json:"status"`
	ErrorType string `json:"errorType"`
	Error     string `json:"error"`
}

type Health struct {
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
	Status       bool     `json:"status"` // Status 0,1,2 maps to health status green (0), orange(1), red (2)
	HealthyNodes []Health `json:"healthyNodes"`
	FailureNodes []Health `json:"failureNodes"`
}

type HealthSummary struct {
	Status   bool
	Services []string
	Failed   []string
}

type PromQR struct {
	Name  string
	Job   string
	Node  string
	Value int64
}

// Consulhealth representates the health state of consul in the  cluster
type ConsulHealth struct {
	Health                 int // 0,1,2
	ConsulUp               bool
	ConsulRaftPeers        bool
	ConsulSerfMembers      bool
	ConsulRaftLeader       bool
	ConsulHealthNodeStatus bool
}
// GlusterHealth representates the health state of glusterfs in the  cluster
type GlusterHealth struct {
	Health int // 0,1,2
	GlusterUp bool
	GlusterPeersConnected bool
	GlusterSuccessfullyMounted bool
	GlusterMountWriteable bool
}