package main

import (
	"encoding/json"
	"errors"
	"fmt"
	api "github.com/ofesseler/panopticon/promapi"
	"reflect"
	"testing"
)

func TestGetUniqueLinks(t *testing.T) {
	links := []api.Link{
		{Source: "w1", Target: "w2", Value: 2},
		{Source: "w2", Target: "w3", Value: 2},
	}
	linksDirty := []api.Link{
		{Source: "w1", Target: "w2", Value: 1},
		{Source: "w2", Target: "w3", Value: 1},
		{Source: "w2", Target: "w1", Value: 1},
		{Source: "w3", Target: "w2", Value: 1},
		{Source: "w2", Target: "w2", Value: 1},
	}

	fn := GetUniqueLinks(linksDirty)
	if !reflect.DeepEqual(fn, links) {
		obj, _ := json.MarshalIndent(fn, "", "  ")
		fmt.Println(string(obj))
		t.Error("not equal")
	}
}

func TestCheckPromResponse(t *testing.T) {

	scrOk := api.StatusCheckReceived{Status: "success"}
	scrOk.Data.ResultType = "vector"
	result := checkPromResponse(scrOk)
	if result == false {
		t.Errorf("checkprom respone wasn't successful with struct %v and returned %v", scrOk, result)
	}
	scrEmpty := api.StatusCheckReceived{}
	result = checkPromResponse(scrEmpty)
	if result == true {
		t.Error("sent empty StatusCheckReceived struct and returned true. Shuld be false.")
	}

	scrResultTypeEmpty := api.StatusCheckReceived{Status: "success"}
	result = checkPromResponse(scrResultTypeEmpty)
	if result == true {
		t.Error("ResultType was empty and checkPromResponse returned true")
	}

	scrResultTypeNotVector := api.StatusCheckReceived{Status: "success"}
	scrResultTypeNotVector.Data.ResultType = "novector"
	result = checkPromResponse(scrResultTypeNotVector)
	if result == true {
		t.Error("ResultType was NOT 'vector' and checkPromResponse returned true")
	}
}

func TestCheckerr(t *testing.T) {
	checkerr(nil)
	err := errors.New("type")
	checkerr(err)

}
