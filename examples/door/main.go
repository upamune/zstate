package main

import (
	"context"
	"fmt"

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

func main() {
	doorBuilder := zstate.NewStateMachineBuilder[DoorState, DoorEvent]()
	door, _ := doorBuilder.
		AddState(Closed).
		AddState(Open).
		AddState(Locked).
		AddTransition(Closed, Open, OpenDoor,
			zstate.WithBefore[DoorState, DoorEvent](func(ctx context.Context, from, to DoorState, event DoorEvent) {
				fmt.Println("[BeforeCallback] Before opening the door")
			}),
			zstate.WithAfter[DoorState, DoorEvent](func(ctx context.Context, from, to DoorState, event DoorEvent) {
				fmt.Println("[AfterCallback] After opening the door")
			}),
		).
		AddTransition(Open, Closed, CloseDoor).
		AddTransition(Closed, Locked, LockDoor, zstate.WithGuard[DoorState, DoorEvent](func(ctx context.Context, from, to DoorState, event DoorEvent) bool {
			return true
		})).
		AddTransition(Locked, Closed, UnlockDoor).
		Build()

	ctx := context.Background()

	newState, err := door.Trigger(ctx, Closed, OpenDoor)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("New state: %v\n", newState)

	diagram, err := zstate.GenerateDiagram(door, zstate.MermaidFormat, Closed)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Diagram: \n%v\n", diagram)
}
