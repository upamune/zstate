package main

import (
	"context"
	"fmt"
	"log"

	"github.com/upamune/zstate"
)

type PlayerState string

const (
	Stopped PlayerState = "Stopped"
	Playing PlayerState = "Playing"
	Paused  PlayerState = "Paused"
)

type PlayerEvent string

const (
	Play  PlayerEvent = "Play"
	Pause PlayerEvent = "Pause"
	Stop  PlayerEvent = "Stop"
	Next  PlayerEvent = "Next"
	Prev  PlayerEvent = "Previous"
)

type MusicPlayer struct {
	stateMachine *zstate.StateMachine[PlayerState, PlayerEvent]
	currentTrack int
	currentState PlayerState
}

func NewMusicPlayer() (*MusicPlayer, error) {
	builder := zstate.NewStateMachineBuilder[PlayerState, PlayerEvent]()
	sm, err := builder.
		AddState(Stopped).
		AddState(Playing).
		AddState(Paused).
		AddTransition(Stopped, Playing, Play, zstate.WithBefore[PlayerState, PlayerEvent](logTransition)).
		AddTransition(Playing, Paused, Pause, zstate.WithBefore[PlayerState, PlayerEvent](logTransition)).
		AddTransition(Paused, Playing, Play, zstate.WithBefore[PlayerState, PlayerEvent](logTransition)).
		AddTransition(Playing, Stopped, Stop, zstate.WithBefore[PlayerState, PlayerEvent](logTransition)).
		AddTransition(Paused, Stopped, Stop, zstate.WithBefore[PlayerState, PlayerEvent](logTransition)).
		AddTransition(Playing, Playing, Next, zstate.WithBefore[PlayerState, PlayerEvent](logTransition)).
		AddTransition(Playing, Playing, Prev, zstate.WithBefore[PlayerState, PlayerEvent](logTransition)).
		Build()

	if err != nil {
		return nil, fmt.Errorf("failed to build state machine: %w", err)
	}

	return &MusicPlayer{
		stateMachine: sm,
		currentTrack: 0,
		currentState: Stopped,
	}, nil
}

func logTransition(ctx context.Context, from, to PlayerState, event PlayerEvent) {
	log.Printf("Transition: %v -> %v (Event: %v)", from, to, event)
}

func (mp *MusicPlayer) Trigger(event PlayerEvent) error {
	ctx := context.Background()
	currentState := mp.GetCurrentState()
	newState, err := mp.stateMachine.Trigger(ctx, currentState, event)
	if err != nil {
		return fmt.Errorf("failed to trigger event %v: %w", event, err)
	}

	mp.handleSideEffects(event)
	log.Printf("New state: %v", newState)
	mp.currentState = newState
	return nil
}

func (mp *MusicPlayer) handleSideEffects(event PlayerEvent) {
	switch event {
	case Next:
		mp.currentTrack++
		log.Printf("Moved to next track: %d", mp.currentTrack)
	case Prev:
		if mp.currentTrack > 0 {
			mp.currentTrack--
			log.Printf("Moved to previous track: %d", mp.currentTrack)
		}
	}
}

func (mp *MusicPlayer) GetCurrentState() PlayerState {
	return mp.currentState
}

func main() {
	player, err := NewMusicPlayer()
	if err != nil {
		log.Fatalf("Failed to create music player: %v", err)
	}

	// Simulate user interactions
	actions := []PlayerEvent{Play, Pause, Play, Next, Prev, Stop}

	for _, action := range actions {
		fmt.Printf("Triggering event: %v\n", action)
		err := player.Trigger(action)
		if err != nil {
			log.Printf("Error: %v", err)
		}
		fmt.Printf("Current state: %v\n\n", player.GetCurrentState())
	}

	// Generate and print the state diagram
	diagram, err := zstate.GenerateDiagram(player.stateMachine, zstate.MermaidFormat, player.GetCurrentState())
	if err != nil {
		log.Fatalf("Failed to generate diagram: %v", err)
	}
	fmt.Println("State Machine Diagram:")
	fmt.Println(diagram)
}
