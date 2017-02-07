package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/looplab/fsm"
)

const (
	// CURRENT is name of current state event
	CURRENT = "current"
	// WARNING is name of warning state event
	WARNING = "warning"
	// FATAL is name of fatal state event
	FATAL = "fatal"
	// RESOLV is name of resolv state event
	RESOLV = "resolv"
)

// Health is cluster health fsm
type Health struct {
	To  string
	FSM *fsm.FSM
}

// State is current cluster state information
type State struct {
	Success bool   `json:"success"`
	Request string `json:"request"`
	Current string `json:"current"`
	Last    string `json:"last"`
	Message string `json:"message"`
}

// NewHealth creates a instance of cluster fsm
func NewHealth(to string) *Health {
	h := &Health{
		To: to,
	}

	h.FSM = fsm.NewFSM("red",
		fsm.Events{
			{Name: "warning", Src: []string{"green"}, Dst: "orange"},
			{Name: "warning", Src: []string{"orange"}, Dst: "orange"},
			{Name: "resolv", Src: []string{"red"}, Dst: "orange"},
			{Name: "resolv", Src: []string{"orange"}, Dst: "green"},
			{Name: "fatal", Src: []string{"green"}, Dst: "red"},
			{Name: "fatal", Src: []string{"orange"}, Dst: "red"},
		},
		fsm.Callbacks{
			"enter_state": func(e *fsm.Event) { h.enterState(e) },
		},
	)
	return h
}

func (h *Health) enterState(e *fsm.Event) {
	log.Infof("The health of %s is %s\n", h.To, e.Dst)
}

func (h *Health) doDegrade(e *fsm.Event) {
	log.Infof("Degrade from %s to %s", e.FSM.Current(), e.FSM.Can("fatal"))
}
