package zstate

import (
	"fmt"
)

// StateError represents an error related to state operations
type StateError[S comparable] struct {
	State S
	Msg   string
}

func (e *StateError[S]) Error() string {
	return fmt.Sprintf("state error: %s (state: %v)", e.Msg, e.State)
}

// TransitionError represents an error related to transition operations
type TransitionError[S, E comparable] struct {
	From  S
	To    S
	Event E
	Msg   string
}

func (e *TransitionError[S, E]) Error() string {
	return fmt.Sprintf("transition error: %s (from: %v, to: %v, event: %v)", e.Msg, e.From, e.To, e.Event)
}

// GuardError represents an error when a guard condition is not met
type GuardError[S, E comparable] struct {
	From  S
	To    S
	Event E
}

func (e *GuardError[S, E]) Error() string {
	return fmt.Sprintf("guard error: condition not met (from: %v, to: %v, event: %v)", e.From, e.To, e.Event)
}

// NoTransitionError represents an error when no transition is found
type NoTransitionError[S, E comparable] struct {
	From  S
	Event E
}

func (e *NoTransitionError[S, E]) Error() string {
	return fmt.Sprintf("no transition error: no transition found (from: %v, event: %v)", e.From, e.Event)
}
