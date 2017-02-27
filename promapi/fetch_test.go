package promapi

import (
	"fmt"
	"testing"
)

type testPair struct {
	f        Fetcher
	expInt   int
	expInt64 int64
	expBool  bool
}

type statusOK struct {
	status string
}

func (fetcher statusOK) PromQuery(query, promHost string) (StatusCheckReceived, error) {
	var scr = StatusCheckReceived{Status: fetcher.status}
	return scr, nil
}

func TestFetcher_PromQuery(t *testing.T) {
	var fok Fetcher = statusOK{status: "ok"}
	scr, err := fok.PromQuery("up", "asd")
	if !(scr.Status == "ok" && err == nil) {
		t.Fatalf("status is %v instead of 'ok' and err was expected 'nil' and is %v", scr.Status, err)
	}

	var fasd Fetcher = statusOK{status: "asd"}
	scr, _ = fasd.PromQuery("up", "asd")
	if scr.Status != "asd" {
		t.Fatalf("scr.Status was 'asd' not %v", scr.Status)
	}

}

type ConsulTest struct {
	Total        int
	Failed       int
	SuccessValue string
	FailValue    string
}

func (f ConsulTest) PromQuery(query, promHost string) (StatusCheckReceived, error) {
	scr := StatusCheckReceived{
		Status: "success",
	}
	scr.Data.ResultType = "vector"
	result := make([]Result, f.Total)

	for i := 0; i < f.Total; i++ {
		var values = make([]interface{}, 2)
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

func TestFetchServiceUp(t *testing.T) {
	var test = []testPair{
		{f: ConsulTest{Total: 3, Failed: 0, SuccessValue: "1", FailValue: "0"}, expBool: true},
		{f: ConsulTest{Total: 3, Failed: 1, SuccessValue: "1", FailValue: "0"}, expBool: false},
		{f: ConsulTest{Total: 3, Failed: 2, SuccessValue: "1", FailValue: "0"}, expBool: false},
	}
	for _, p := range test {
		healthStatus, err := FetchServiceUp(p.f, ConsulUp, "FetchServiceUp")
		if err != nil {
			t.Fatal(err)
		}
		if healthStatus.Status != p.expBool {
			t.FailNow()
		}
	}
}

func TestFetchPromGauge(t *testing.T) {
	// expInt: in this case is for no. of results
	var test = []testPair{
		{f: ConsulTest{Total: 3, Failed: 0, SuccessValue: "3", FailValue: "0"}, expInt64: 3, expInt: 3},
		{f: ConsulTest{Total: 3, Failed: 1, SuccessValue: "2", FailValue: "2"}, expInt64: 2, expInt: 3},
		{f: ConsulTest{Total: 3, Failed: 2, SuccessValue: "1", FailValue: "1"}, expInt64: 1, expInt: 3},
	}

	for _, p := range test {
		h, err := FetchPromGauge(p.f, "PromGauge", "up")
		if err != nil {
			t.Error(err)
		}

		if len(h) != p.expInt {
			t.Errorf("Expected %v instances but only %v results", p.expInt, p.expInt64)
		}
		for _, v := range h {
			if v.Value != p.expInt64 {
				t.Errorf("Expected peer count was %v and got %v", p.expInt64, v.Value)
			}
		}
	}
}
