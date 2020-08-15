package state

import (
	"fmt"
	"reflect"

	"github.com/go-playground/log/v7"
)

// State is the object that performs actions while it it active
type State interface {
	// Activate starts the state's execution
	Activate(stop <-chan bool, manager *Manager)

	// AllowedFrom returns a slice of states that are allowed to transition into this state
	AllowedFrom() []State

	// Done returns a channel that will close once the state processing has finished
	Done() <-chan bool

	// Resume resumes a states execution
	Resume(stop <-chan bool, manager *Manager)
}

// NewManager initializes and returns a new manager instance
func NewManager(stop <-chan bool) *Manager {
	m := &Manager{nil, make(chan State), make(chan State), make(chan chan State)}
	m.runner(stop)
	return m
}

// Manager manages the execution of states
type Manager struct {
	currentState            State
	incomingState           chan State
	incomingResumeState     chan State
	incomingCurrentStateReq chan chan State
}

// TransitionTo transitions into the given state
func (m *Manager) TransitionTo(state State) error {
	log.Infof("transitioning into state %T", state)
	if m.currentState != nil {
		for _, s := range m.currentState.AllowedFrom() {
			if reflect.TypeOf(s) == reflect.TypeOf(state) {
				m.incomingState <- state
				return nil
			}
		}
		return fmt.Errorf("cannot transition into state: state doesn't allow transitions from %T state type", state)
	}
	m.incomingState <- state
	return nil
}

// Resume transitions into the given state without activating the processing
func (m *Manager) Resume(state State) error {
	log.Infof("resuming state %T", state)
	m.incomingResumeState <- state
	return nil
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
				m.currentState = state
				go state.Activate(stop, m)
			case state := <-m.incomingResumeState:
				m.currentState = state
				go state.Resume(stop, m)
			case ch := <-m.incomingCurrentStateReq:
				ch <- m.currentState
				close(ch)
			case <-stop:
				return
			}
		}
	}(stop)
}
