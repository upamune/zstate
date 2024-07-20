// Package zstate provides a generic and flexible state machine implementation.
package zstate

import (
	"context"
	"fmt"
	"sync"
)

// BeforeTransitionEvent is an interface for events that occur before a state transition.
type BeforeTransitionEvent interface {
	BeforeTransition(ctx context.Context)
}

// AfterTransitionEvent is an interface for events that occur after a state transition.
type AfterTransitionEvent interface {
	AfterTransition(ctx context.Context)
}

// Guard is a function type that determines if a transition is allowed.
// It receives the context, current state, target state, and the event triggering the transition.
type Guard[S, E comparable] func(ctx context.Context, from S, to S, event E) bool

// transition represents a transition in the state machine.
type transition[S, E comparable] struct {
	from  S
	to    S
	event E
	guard Guard[S, E]
}

// StateMachine represents the main state machine entity with generic state type S and event type E.
type StateMachine[S, E comparable] struct {
	states       map[S]struct{}
	transitions  map[S]map[E]transition[S, E]
	currentState S
	mu           sync.RWMutex
}

// StateMachineBuilder is the interface for building a state machine.
type StateMachineBuilder[S, E comparable] interface {
	// AddState adds a new state to the state machine.
	AddState(s S) StateMachineBuilder[S, E]
	// SetInitialState sets the initial state of the state machine.
	SetInitialState(s S) StateMachineBuilder[S, E]
	// AddTransition adds a new transition to the state machine.
	AddTransition(from, to S, event E, opts ...TransitionOption[S, E]) StateMachineBuilder[S, E]
	// Build finalizes the construction of the state machine.
	Build() (*StateMachine[S, E], error)
}

type stateMachineBuilder[S, E comparable] struct {
	states            map[S]struct{}
	transitions       map[S]map[E]transition[S, E]
	initialState      S
	isInitialStateSet bool
}

// TransitionOption is a function type for configuring transitions.
type TransitionOption[S, E comparable] func(*transition[S, E])

// WithGuard adds a guard function to a transition.
func WithGuard[S, E comparable](guard Guard[S, E]) TransitionOption[S, E] {
	return func(t *transition[S, E]) {
		t.guard = guard
	}
}

// NewStateMachineBuilder creates a new StateMachineBuilder.
func NewStateMachineBuilder[S, E comparable]() StateMachineBuilder[S, E] {
	return &stateMachineBuilder[S, E]{
		states:      make(map[S]struct{}),
		transitions: make(map[S]map[E]transition[S, E]),
	}
}

// AddState adds a new state to the state machine.
func (b *stateMachineBuilder[S, E]) AddState(s S) StateMachineBuilder[S, E] {
	b.states[s] = struct{}{}
	return b
}

// SetInitialState sets the initial state of the state machine.
func (b *stateMachineBuilder[S, E]) SetInitialState(s S) StateMachineBuilder[S, E] {
	b.initialState = s
	b.isInitialStateSet = true
	return b
}

// AddTransition adds a new transition to the state machine.
func (b *stateMachineBuilder[S, E]) AddTransition(from, to S, event E, opts ...TransitionOption[S, E]) StateMachineBuilder[S, E] {
	t := transition[S, E]{
		from:  from,
		to:    to,
		event: event,
	}

	for _, opt := range opts {
		opt(&t)
	}

	if b.transitions[from] == nil {
		b.transitions[from] = make(map[E]transition[S, E])
	}
	b.transitions[from][event] = t
	return b
}

// Build finalizes the construction of the state machine.
func (b *stateMachineBuilder[S, E]) Build() (*StateMachine[S, E], error) {
	if len(b.states) == 0 {
		return nil, fmt.Errorf("state machine must have at least one state")
	}

	if !b.isInitialStateSet {
		return nil, fmt.Errorf("initial state must be set")
	}

	if _, ok := b.states[b.initialState]; !ok {
		return nil, fmt.Errorf("initial state must be a valid state")
	}

	return &StateMachine[S, E]{
		states:       b.states,
		transitions:  b.transitions,
		currentState: b.initialState,
	}, nil
}

// Trigger attempts to transition the state machine based on the given event.
// It returns an error if the transition is not possible.
func (sm *StateMachine[S, E]) Trigger(ctx context.Context, event E) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if transitions, ok := sm.transitions[sm.currentState]; ok {
		if t, ok := transitions[event]; ok {
			if t.guard == nil || t.guard(ctx, sm.currentState, t.to, event) {
				// Check if the event implements BeforeTransitionEvent interface and call BeforeTransition if it does
				if beforeEvent, ok := any(event).(BeforeTransitionEvent); ok {
					beforeEvent.BeforeTransition(ctx)
				}

				sm.currentState = t.to

				// Check if the event implements AfterTransitionEvent interface and call AfterTransition if it does
				if afterEvent, ok := any(event).(AfterTransitionEvent); ok {
					afterEvent.AfterTransition(ctx)
				}

				return nil
			}
		}
	}
	return fmt.Errorf("no valid transition for current state and given event")
}

// GetCurrentState returns the current state of the state machine.
func (sm *StateMachine[S, E]) GetCurrentState() S {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.currentState
}
