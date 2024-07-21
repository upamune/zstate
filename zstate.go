// Package zstate provides a generic and flexible state machine implementation.
package zstate

import (
	"context"
)

// StateMachine represents the state machine entity with generic state type S and event type E
type StateMachine[S, E comparable] struct {
	states      map[S]struct{}
	transitions map[S]map[E]transition[S, E]
}

// transition represents a transition in the state machine
type transition[S, E comparable] struct {
	from   S
	to     S
	event  E
	guard  Guard[S, E]
	before TransitionCallback[S, E]
	after  TransitionCallback[S, E]
}

// Guard is a function type that determines if a transition is allowed
type Guard[S, E comparable] func(ctx context.Context, from, to S, event E) bool

// TransitionCallback is a function type for before and after transition callbacks
type TransitionCallback[S, E comparable] func(ctx context.Context, from, to S, event E)

// StateMachineBuilder is the interface for building a state machine
type StateMachineBuilder[S, E comparable] interface {
	AddState(s S) StateMachineBuilder[S, E]
	AddTransition(from, to S, event E, opts ...TransitionOption[S, E]) StateMachineBuilder[S, E]
	Build() (*StateMachine[S, E], error)
}

type stateMachineBuilder[S, E comparable] struct {
	states      map[S]struct{}
	transitions map[S]map[E]transition[S, E]
}

// TransitionOption is a function type for configuring transitions
type TransitionOption[S, E comparable] func(*transition[S, E])

// WithGuard adds a guard function to a transition
func WithGuard[S, E comparable](guard Guard[S, E]) TransitionOption[S, E] {
	return func(t *transition[S, E]) {
		t.guard = guard
	}
}

// WithBefore adds a before callback to a transition
func WithBefore[S, E comparable](callback TransitionCallback[S, E]) TransitionOption[S, E] {
	return func(t *transition[S, E]) {
		t.before = callback
	}
}

// WithAfter adds an after callback to a transition
func WithAfter[S, E comparable](callback TransitionCallback[S, E]) TransitionOption[S, E] {
	return func(t *transition[S, E]) {
		t.after = callback
	}
}

// NewStateMachineBuilder creates a new StateMachineBuilder
func NewStateMachineBuilder[S, E comparable]() StateMachineBuilder[S, E] {
	return &stateMachineBuilder[S, E]{
		states:      make(map[S]struct{}),
		transitions: make(map[S]map[E]transition[S, E]),
	}
}

// AddState adds a new state to the state machine
func (b *stateMachineBuilder[S, E]) AddState(s S) StateMachineBuilder[S, E] {
	b.states[s] = struct{}{}
	return b
}

// AddTransition adds a new transition to the state machine
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

// Build finalizes the construction of the state machine
func (b *stateMachineBuilder[S, E]) Build() (*StateMachine[S, E], error) {
	if len(b.states) == 0 {
		return nil, &StateError[S]{Msg: "state machine must have at least one state"}
	}

	return &StateMachine[S, E]{
		states:      b.states,
		transitions: b.transitions,
	}, nil
}

// Trigger attempts to perform a transition based on the given event
func (sm *StateMachine[S, E]) Trigger(ctx context.Context, currentState S, event E) (S, error) {
	t, ok := sm.transitions[currentState][event]
	if !ok {
		return currentState, &NoTransitionError[S, E]{From: currentState, Event: event}
	}

	if t.guard != nil && !t.guard(ctx, currentState, t.to, event) {
		return currentState, &GuardError[S, E]{From: currentState, To: t.to, Event: event}
	}

	if t.before != nil {
		t.before(ctx, currentState, t.to, event)
	}

	if t.after != nil {
		defer t.after(ctx, currentState, t.to, event)
	}

	return t.to, nil
}
