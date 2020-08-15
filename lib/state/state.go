package state

import (
	"errors"
	"reflect"
)

// State is the object that performs actions while it it active
type State interface {
	// Activate starts the state's execution
	Activate(stop <-chan bool, manager *Manager)

	// AllowedFrom returns a slice of states that are allowed to transition into this state
	AllowedFrom() []State
}

// NewManager initializes and returns a new manager instance
func NewManager(stop <-chan bool) *Manager {
	m := &Manager{nil, make(chan State), make(chan chan State)}
	m.runner(stop)
	return m
}

// Manager manages the execution of states
type Manager struct {
	currentState            State
	incomingState           chan State
	incomingCurrentStateReq chan chan State
}

// TransitionTo transitions into the given state
func (m *Manager) TransitionTo(state State) error {
	for _, s := range state.AllowedFrom() {
		if reflect.TypeOf(s) == reflect.TypeOf(state) {
			m.incomingState <- state
			return nil
		}
	}
	return errors.New("cannot transition into state: ")
}

// CurrentState returns the current executing state
func (m *Manager) CurrentState() State {
	// Make a request channel
	ch := make(chan State)

	// Send the request
	m.incomingCurrentStateReq <- ch

	// Return the state
	return <-ch
}

func (m *Manager) runner(stop <-chan bool) {
	go func(stop <-chan bool) {
		for {
			select {
			case state := <-m.incomingState:
				go state.Activate(stop, m)
			case ch := <-m.incomingCurrentStateReq:
				ch <- m.currentState
				close(ch)
			case <-stop:
				return
			}
		}
	}(stop)
}
