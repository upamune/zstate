package zstate_test

import (
	"context"
	"fmt"
	"log"

	"github.com/upamune/zstate"
)

type PlayerState int

const (
	Stopped PlayerState = iota
	Playing
	Paused
)

func (s PlayerState) String() string {
	return [...]string{"Stopped", "Playing", "Paused"}[s]
}

type PlayerEvent string

const (
	Play  PlayerEvent = "Play"
	Pause PlayerEvent = "Pause"
	Stop  PlayerEvent = "Stop"
)

type MusicPlayer struct {
	stateMachine *zstate.StateMachine[PlayerState, PlayerEvent]
	currentState PlayerState
}

func NewMusicPlayer() (*MusicPlayer, error) {
	builder := zstate.NewStateMachineBuilder[PlayerState, PlayerEvent]()
	sm, err := builder.
		AddState(Stopped).
		AddState(Playing).
		AddState(Paused).
		AddTransition(Stopped, Playing, Play).
		AddTransition(Playing, Paused, Pause).
		AddTransition(Paused, Playing, Play).
		AddTransition(Playing, Stopped, Stop).
		AddTransition(Paused, Stopped, Stop).
		Build()

	if err != nil {
		return nil, fmt.Errorf("failed to build state machine: %w", err)
	}

	return &MusicPlayer{
		stateMachine: sm,
	}, nil
}

func (mp *MusicPlayer) Trigger(event PlayerEvent) error {
	ctx := context.Background()
	currentState := mp.GetCurrentState()
	nextState, err := mp.stateMachine.Trigger(ctx, currentState, event)
	if err != nil {
		return err
	}
	mp.currentState = nextState
	return nil
}

func (mp *MusicPlayer) GetCurrentState() PlayerState {
	return mp.currentState
}

func Example() {
	player, err := NewMusicPlayer()
	if err != nil {
		log.Fatalf("Failed to create music player: %v", err)
	}

	fmt.Printf("Initial state: %v\n", player.GetCurrentState())

	actions := []PlayerEvent{Play, Pause, Play, Stop}

	for _, action := range actions {
		err := player.Trigger(action)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		} else {
			fmt.Printf("Action: %v, New state: %v\n", action, player.GetCurrentState())
		}
	}

	diagram, err := zstate.GenerateDiagram(player.stateMachine, zstate.MermaidFormat, Stopped)
	if err != nil {
		log.Fatalf("Failed to generate diagram: %v", err)
	}
	fmt.Println("Diagram:")
	fmt.Println(diagram)

	// Output:
	// Initial state: Stopped
	// Action: Play, New state: Playing
	// Action: Pause, New state: Paused
	// Action: Play, New state: Playing
	// Action: Stop, New state: Stopped
	// Diagram:
	// stateDiagram-v2
	//     Paused
	//     Playing
	//     Stopped : [*] Stopped
	//     Paused --> Playing : Play
	//     Paused --> Stopped : Stop
	//     Playing --> Paused : Pause
	//     Playing --> Stopped : Stop
	//     Stopped --> Playing : Play
}
