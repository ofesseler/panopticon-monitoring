package main

import (
	"testing"
	"reflect"
	"encoding/json"
	"fmt"
)

func TestGetUniqueLinks(t *testing.T) {
	links := []Link{
		{Source:"w1", Target:"w2", Value: 2},
		{Source:"w2", Target:"w3", Value: 2},
	}
	links_dirty := []Link{
		{Source:"w1", Target:"w2", Value: 1},
		{Source:"w2", Target:"w3", Value: 1},
		{Source:"w2", Target:"w1", Value: 1},
		{Source:"w3", Target:"w2", Value: 1},
		{Source:"w2", Target:"w2", Value: 1},
	}

	fn := GetUniqueLinks(links_dirty)
	if !reflect.DeepEqual(fn, links) {
		obj, _ := json.MarshalIndent(fn,"", "  ")
		fmt.Println(string(obj))
		t.Error("not equal")
	}
}