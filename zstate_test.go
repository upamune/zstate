package zstate_test

import (
	"context"
	"testing"

	"github.com/upamune/zstate"
)

type DoorState int

const (
	Closed DoorState = iota
	Open
	Locked
)

type DoorEvent int

const (
	OpenDoor DoorEvent = iota
	CloseDoor
	LockDoor
	UnlockDoor
)

func TestDoorStateMachine(t *testing.T) {
	var isLocked bool

	doorBuilder := zstate.NewStateMachineBuilder[DoorState, DoorEvent]()
	door, err := doorBuilder.
		AddState(Closed).
		AddState(Open).
		AddState(Locked).
		SetInitialState(Closed).
		AddTransition(Closed, Open, OpenDoor).
		AddTransition(Open, Closed, CloseDoor).
		AddTransition(Closed, Locked, LockDoor, zstate.WithGuard[DoorState, DoorEvent](func(ctx context.Context, from, to DoorState, event DoorEvent) bool {
			return !isLocked
		})).
		AddTransition(Locked, Closed, UnlockDoor).
		Build()

	if err != nil {
		t.Fatalf("Error building state machine: %v", err)
	}

	ctx := context.Background()

	t.Run("Initial State", func(t *testing.T) {
		if door.GetCurrentState() != Closed {
			t.Errorf("Expected initial state to be Closed, got %v", door.GetCurrentState())
		}
	})

	t.Run("Valid Transition", func(t *testing.T) {
		err := door.Trigger(ctx, OpenDoor)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if door.GetCurrentState() != Open {
			t.Errorf("Expected state to be Open, got %v", door.GetCurrentState())
		}
	})

	t.Run("Invalid Transition", func(t *testing.T) {
		err := door.Trigger(ctx, LockDoor) // Can't lock an open door
		if err == nil {
			t.Error("Expected error for invalid transition, got nil")
		}
	})

	t.Run("Guard Function", func(t *testing.T) {
		_ = door.Trigger(ctx, CloseDoor) // Close the door first
		isLocked = false
		err := door.Trigger(ctx, LockDoor)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if door.GetCurrentState() != Locked {
			t.Errorf("Expected state to be Locked, got %v", door.GetCurrentState())
		}

		isLocked = true
		err = door.Trigger(ctx, UnlockDoor)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if door.GetCurrentState() != Closed {
			t.Errorf("Expected state to be Closed, got %v", door.GetCurrentState())
		}

		err = door.Trigger(ctx, LockDoor)
		if err == nil {
			t.Error("Expected error due to guard function, got nil")
		}
		if door.GetCurrentState() != Closed {
			t.Errorf("Expected state to remain Closed, got %v", door.GetCurrentState())
		}
	})

	t.Run("Multiple Transitions", func(t *testing.T) {
		transitions := []DoorEvent{OpenDoor, CloseDoor, LockDoor, UnlockDoor}
		expectedStates := []DoorState{Open, Closed, Locked, Closed}

		for i, transition := range transitions {
			isLocked = false // Reset lock for each transition
			err := door.Trigger(ctx, transition)
			if err != nil {
				t.Errorf("Unexpected error on transition %v: %v", transition, err)
			}
			if door.GetCurrentState() != expectedStates[i] {
				t.Errorf("Expected state to be %v, got %v", expectedStates[i], door.GetCurrentState())
			}
		}
	})
}

// DoorEventWithCallback implements both BeforeTransitionEvent and AfterTransitionEvent
type DoorEventWithCallback struct {
	DoorEvent
	BeforeCalled bool
	AfterCalled  bool
}

var (
	open DoorEventWithCallback = DoorEventWithCallback{DoorEvent: OpenDoor}
)

func (e *DoorEventWithCallback) BeforeTransition(ctx context.Context) {
	e.BeforeCalled = true
}

func (e *DoorEventWithCallback) AfterTransition(ctx context.Context) {
	e.AfterCalled = true
}

func TestBeforeAfterTransitionEvents(t *testing.T) {
	doorBuilder := zstate.NewStateMachineBuilder[DoorState, *DoorEventWithCallback]()
	door, err := doorBuilder.
		AddState(Closed).
		AddState(Open).
		SetInitialState(Closed).
		AddTransition(Closed, Open, &open).
		Build()
	if err != nil {
		t.Fatalf("failed to build state machine: %v", err)
	}

	ctx := context.Background()
	event := &open
	if err := door.Trigger(ctx, event); err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if !event.BeforeCalled {
		t.Error("BeforeTransition was not called")
	}
	if !event.AfterCalled {
		t.Error("AfterTransition was not called")
	}
	if door.GetCurrentState() != Open {
		t.Errorf("Expected state to be Open, got %v", door.GetCurrentState())
	}
}

// OrderState represents the possible states of an order
type OrderState int

const (
	Created OrderState = iota
	PaymentPending
	Paid
	Shipped
	Delivered
	Cancelled
)

// OrderEvent represents the possible events that can occur in the order process
type OrderEvent string

const (
	SubmitPayment  OrderEvent = "SubmitPayment"
	ConfirmPayment OrderEvent = "ConfirmPayment"
	ShipOrder      OrderEvent = "ShipOrder"
	DeliverOrder   OrderEvent = "DeliverOrder"
	CancelOrder    OrderEvent = "CancelOrder"
)

func TestOrderStateMachine(t *testing.T) {
	var orderAmount float64
	var isStockAvailable bool

	orderBuilder := zstate.NewStateMachineBuilder[OrderState, OrderEvent]()
	order, err := orderBuilder.
		AddState(Created).
		AddState(PaymentPending).
		AddState(Paid).
		AddState(Shipped).
		AddState(Delivered).
		AddState(Cancelled).
		SetInitialState(Created).
		AddTransition(Created, PaymentPending, SubmitPayment).
		AddTransition(PaymentPending, Paid, ConfirmPayment,
			zstate.WithGuard[OrderState, OrderEvent](func(ctx context.Context, from, to OrderState, event OrderEvent) bool {
				return orderAmount >= 100.0 // Assuming minimum order amount is 100
			})).
		AddTransition(Paid, Shipped, ShipOrder,
			zstate.WithGuard[OrderState, OrderEvent](func(ctx context.Context, from, to OrderState, event OrderEvent) bool {
				return isStockAvailable
			})).
		AddTransition(Shipped, Delivered, DeliverOrder).
		AddTransition(Created, Cancelled, CancelOrder).
		AddTransition(PaymentPending, Cancelled, CancelOrder).
		AddTransition(Paid, Cancelled, CancelOrder).
		Build()

	if err != nil {
		t.Fatalf("Error building state machine: %v", err)
	}

	ctx := context.Background()

	t.Run("Initial State", func(t *testing.T) {
		if order.GetCurrentState() != Created {
			t.Errorf("Expected initial state to be Created, got %v", order.GetCurrentState())
		}
	})

	t.Run("Valid Transition Sequence", func(t *testing.T) {
		orderAmount = 100.0
		isStockAvailable = true

		events := []OrderEvent{SubmitPayment, ConfirmPayment, ShipOrder, DeliverOrder}
		expectedStates := []OrderState{PaymentPending, Paid, Shipped, Delivered}

		for i, event := range events {
			err := order.Trigger(ctx, event)
			if err != nil {
				t.Errorf("Unexpected error on event %v: %v", event, err)
			}
			if order.GetCurrentState() != expectedStates[i] {
				t.Errorf("Expected state to be %v, got %v", expectedStates[i], order.GetCurrentState())
			}
		}
	})

	t.Run("Guard Function - Insufficient Payment", func(t *testing.T) {
		order, _ = orderBuilder.Build() // Reset the state machine
		orderAmount = 50.0

		_ = order.Trigger(ctx, SubmitPayment)
		err := order.Trigger(ctx, ConfirmPayment)
		if err == nil {
			t.Error("Expected error due to insufficient payment, got nil")
		}
		if order.GetCurrentState() != PaymentPending {
			t.Errorf("Expected state to remain PaymentPending, got %v", order.GetCurrentState())
		}
	})

	t.Run("Guard Function - Out of Stock", func(t *testing.T) {
		order, _ = orderBuilder.Build() // Reset the state machine
		orderAmount = 100.0
		isStockAvailable = false

		_ = order.Trigger(ctx, SubmitPayment)
		_ = order.Trigger(ctx, ConfirmPayment)
		err := order.Trigger(ctx, ShipOrder)
		if err == nil {
			t.Error("Expected error due to out of stock, got nil")
		}
		if order.GetCurrentState() != Paid {
			t.Errorf("Expected state to remain Paid, got %v", order.GetCurrentState())
		}
	})

	t.Run("Cancel Order", func(t *testing.T) {
		order, _ = orderBuilder.Build() // Reset the state machine

		_ = order.Trigger(ctx, SubmitPayment)
		_ = order.Trigger(ctx, ConfirmPayment)

		err := order.Trigger(ctx, CancelOrder)
		if err != nil {
			t.Errorf("Unexpected error on cancelling order: %v", err)
		}
		if order.GetCurrentState() != Cancelled {
			t.Errorf("Expected state to be Cancelled, got %v", order.GetCurrentState())
		}
	})
}
