package main

import (
	api "github.com/ofesseler/panopticon/promapi"
	"testing"

	"fmt"
)

type testPair struct {
	f        api.Fetcher
	expInt   int
	expInt64 int64
	expBool  bool
}

type ConsulTest struct {
	Total        int
	Failed       int
	SuccessValue string
	FailValue    string
}

func (f ConsulTest) PromQuery(query, promHost string) (api.StatusCheckReceived, error) {
	scr := api.StatusCheckReceived{
		Status: "success",
	}
	scr.Data.ResultType = "vector"
	result := make([]api.Result, f.Total)

	for i := 0; i < f.Total; i++ {
		var values []interface{} = make([]interface{}, 2)
		if i < f.Total-f.Failed {
			values[1] = f.SuccessValue
		} else {
			values[1] = f.FailValue
		}
		result[i].Value = values
		result[i].Metric.Name = fmt.Sprintf("name_%v", i)
		result[i].Metric.Instance = fmt.Sprintf("http://%v", promHost)
		result[i].Metric.Job = fmt.Sprintf("job_%v", i)
	}
	scr.Data.Result = result
	return scr, nil
}

func TestProcessGlusterHealthSummary(t *testing.T) {
	api.ClusterNodeCount = 2
	var test = []testPair{
		{f: ConsulTest{Total: 2, Failed: 0, SuccessValue: "1", FailValue: "0"}, expInt64: 1},
	}

	for _, p := range test {
		s, err := ProcessGlusterHealthSummary(p.f, "ProcessGlusterHealthSummary")
		if err != nil {
			t.Error(err)
		}
		if !s.GlusterUp {
			t.Error("GlusterUp failed")
		}
		if !s.GlusterPeersConnected {
			t.Error("Peer connection failed")
		}
		// TODO
		/*
			if !s.GlusterSuccessfullyMounted {
				t.Error("Mount failed")
			}
			if !s.GlusterMountWriteable {
				t.Error("Gluster write on mount failed")
			}*/
	}
}

func TestComputeCountersFromPromQRs(t *testing.T) {
	api.ClusterNodeCount = 3
	type GaugeToState struct {
		Status api.ClusterStatus
		Arr    []api.PromQR
	}
	test := []GaugeToState{
		{Status: api.HEALTHY, Arr: []api.PromQR{{Node: "node1", Value: 3}, {Node: "node2", Value: 3}, {Node: "node3", Value: 3}}},
		{Status: api.WARNING, Arr: []api.PromQR{{Node: "node1", Value: 2}, {Node: "node2", Value: 2}, {Node: "node3", Value: 0}}},
		{Status: api.WARNING, Arr: []api.PromQR{{Node: "node1", Value: 2}, {Node: "node2", Value: 0}, {Node: "node3", Value: 2}}},
		{Status: api.CRITICAL, Arr: []api.PromQR{{Node: "node1", Value: 1}, {Node: "node2", Value: 0}, {Node: "node3", Value: 0}}},
		{Status: api.WARNING, Arr: []api.PromQR{{Node: "node1", Value: 3}, {Node: "node2", Value: 3}, {Node: "node3", Value: 0}}},
		{Status: api.CRITICAL, Arr: []api.PromQR{{Node: "node1", Value: 1}, {Node: "node2", Value: 1}, {Node: "node3", Value: 1}}},
	}

	for _, b := range test {

		a := computeCountersFromPromQRs(b.Arr)
		if a != b.Status {
			t.Errorf("For element: %v expected %v actual is : %v", b.Arr, b.Status, a)
		}
	}

}

func TestQuorumRate_Rater(t *testing.T) {
	type RateTester struct {
		v interface{}
		r interface{}
		e api.ClusterStatus
	}
	tests := []RateTester{
		{v: 3, r: 3, e: api.HEALTHY},
		{v: 2, r: 3, e: api.WARNING},
		{v: 1, r: 3, e: api.CRITICAL},
		{v: 0, r: 3, e: api.CRITICAL},
		{v: 5, r: 7, e: api.WARNING},
	}
	qr := QuorumRate{}

	for _, test := range tests {
		status, _ := qr.Rater(test.v, test.r)
		if status != test.e {
			t.Error("Wrong state")
		}
	}

	_, err := qr.Rater("", 0)
	if err == nil {
		t.Error("ecpected conversion error")
	}
	_, err = qr.Rater(1, "")
	if err == nil {
		t.Error("ecpected conversion error")
	}
}
