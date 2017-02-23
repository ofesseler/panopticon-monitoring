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
