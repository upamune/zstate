package zstate_test

import (
	"context"
	"errors"
	"testing"

	"github.com/upamune/zstate"
)

type DoorState string

const (
	Closed DoorState = "Closed"
	Open   DoorState = "Open"
	Locked DoorState = "Locked"
)

type DoorEvent string

const (
	OpenDoor   DoorEvent = "OpenDoor"
	CloseDoor  DoorEvent = "CloseDoor"
	LockDoor   DoorEvent = "LockDoor"
	UnlockDoor DoorEvent = "UnlockDoor"
)

func TestStateMachine(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("no states", func(t *testing.T) {
		builder := zstate.NewStateMachineBuilder[DoorState, DoorEvent]()
		_, err := builder.Build()
		if err == nil {
			t.Fatalf("Expected error, got nil")
		}
	})

	t.Run("Basic Transitions", func(t *testing.T) {
		sm := buildDoorStateMachine(t)

		newState, err := sm.Trigger(ctx, Closed, OpenDoor)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if newState != Open {
			t.Errorf("Expected state Open, got %v", newState)
		}

		newState, err = sm.Trigger(ctx, Open, CloseDoor)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if newState != Closed {
			t.Errorf("Expected state Closed, got %v", newState)
		}
	})

	t.Run("No Transition", func(t *testing.T) {
		sm := buildDoorStateMachine(t)

		_, err := sm.Trigger(ctx, Closed, UnlockDoor)
		var noTransitionErr *zstate.NoTransitionError[DoorState, DoorEvent]
		if !errors.As(err, &noTransitionErr) {
			t.Fatalf("Expected NoTransitionError, got %v", err)
		}
	})

	t.Run("WithBefore and WithAfter", func(t *testing.T) {
		var beforeCalled, afterCalled bool
		builder := zstate.NewStateMachineBuilder[DoorState, DoorEvent]()
		sm, err := builder.
			AddState(Closed).
			AddState(Open).
			AddTransition(Closed, Open, OpenDoor,
				zstate.WithBefore[DoorState, DoorEvent](func(ctx context.Context, from, to DoorState, event DoorEvent) {
					beforeCalled = true
				}),
				zstate.WithAfter[DoorState, DoorEvent](func(ctx context.Context, from, to DoorState, event DoorEvent) {
					afterCalled = true
				}),
			).
			Build()

		if err != nil {
			t.Fatalf("Failed to build state machine: %v", err)
		}

		_, err = sm.Trigger(context.Background(), Closed, OpenDoor)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if !beforeCalled {
			t.Error("Before callback was not called")
		}
		if !afterCalled {
			t.Error("After callback was not called")
		}
	})

	t.Run("Guard Condition", func(t *testing.T) {
		builder := zstate.NewStateMachineBuilder[DoorState, DoorEvent]()
		sm, err := builder.
			AddState(Closed).
			AddState(Open).
			AddState(Locked).
			AddTransition(Closed, Locked, LockDoor, zstate.WithGuard[DoorState, DoorEvent](func(ctx context.Context, from, to DoorState, event DoorEvent) bool {
				return false // Always return false to test guard condition
			})).
			Build()

		if err != nil {
			t.Fatalf("Failed to build state machine: %v", err)
		}

		_, err = sm.Trigger(context.Background(), Closed, LockDoor)
		var guardErr *zstate.GuardError[DoorState, DoorEvent]
		if !errors.As(err, &guardErr) {
			t.Fatalf("Expected GuardError, got %T: %v", err, err)
		}
	})
}

func buildDoorStateMachine(t *testing.T) *zstate.StateMachine[DoorState, DoorEvent] {
	t.Helper()

	builder := zstate.NewStateMachineBuilder[DoorState, DoorEvent]()
	sm, err := builder.
		AddState(Closed).
		AddState(Open).
		AddState(Locked).
		AddTransition(Closed, Open, OpenDoor).
		AddTransition(Open, Closed, CloseDoor).
		AddTransition(Closed, Locked, LockDoor, zstate.WithGuard[DoorState, DoorEvent](func(ctx context.Context, from, to DoorState, event DoorEvent) bool {
			return from == Closed // Can only lock when closed
		})).
		AddTransition(Locked, Closed, UnlockDoor).
		Build()

	if err != nil {
		t.Fatalf("Failed to build state machine: %v", err)
	}
	return sm
}
