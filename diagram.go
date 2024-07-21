package zstate

import (
	"fmt"
	"sort"
	"strings"
)

// DiagramFormat represents the format of the generated diagram
type DiagramFormat int

const (
	MermaidFormat DiagramFormat = iota
	DOTFormat
)

// DiagramGenerator is an interface for generating diagrams
type DiagramGenerator interface {
	Generate(states map[string]struct{}, transitions map[string]map[string]string, currentState string) string
}

// MermaidGenerator generates Mermaid diagram
type MermaidGenerator struct{}

func (g *MermaidGenerator) Generate(states map[string]struct{}, transitions map[string]map[string]string, currentState string) string {
	var sb strings.Builder

	sb.WriteString("stateDiagram-v2\n")

	// Sort states for deterministic output
	sortedStates := make([]string, 0, len(states))
	for state := range states {
		sortedStates = append(sortedStates, state)
	}
	sort.Strings(sortedStates)

	for _, state := range sortedStates {
		if state == currentState {
			sb.WriteString(fmt.Sprintf("    %v : [*] %v\n", state, state))
		} else {
			sb.WriteString(fmt.Sprintf("    %v\n", state))
		}
	}

	// Sort transitions for deterministic output
	type transition struct {
		from, to, event string
	}
	sortedTransitions := make([]transition, 0)
	for from, events := range transitions {
		for event, to := range events {
			sortedTransitions = append(sortedTransitions, transition{from, to, event})
		}
	}
	sort.Slice(sortedTransitions, func(i, j int) bool {
		if sortedTransitions[i].from != sortedTransitions[j].from {
			return sortedTransitions[i].from < sortedTransitions[j].from
		}
		if sortedTransitions[i].to != sortedTransitions[j].to {
			return sortedTransitions[i].to < sortedTransitions[j].to
		}
		return sortedTransitions[i].event < sortedTransitions[j].event
	})

	for _, t := range sortedTransitions {
		sb.WriteString(fmt.Sprintf("    %v --> %v : %v\n", t.from, t.to, t.event))
	}

	return sb.String()
}

// DOTGenerator generates DOT diagram
type DOTGenerator struct{}

func (g *DOTGenerator) Generate(states map[string]struct{}, transitions map[string]map[string]string, currentState string) string {
	var sb strings.Builder

	sb.WriteString("digraph StateMachine {\n")

	// Sort states for deterministic output
	sortedStates := make([]string, 0, len(states))
	for state := range states {
		sortedStates = append(sortedStates, state)
	}
	sort.Strings(sortedStates)

	for _, state := range sortedStates {
		if state == currentState {
			sb.WriteString(fmt.Sprintf("    \"%v\" [shape=doublecircle, style=filled, fillcolor=lightblue];\n", state))
		} else {
			sb.WriteString(fmt.Sprintf("    \"%v\" [shape=circle];\n", state))
		}
	}

	// Sort transitions for deterministic output
	type transition struct {
		from, to, event string
	}
	sortedTransitions := make([]transition, 0)
	for from, events := range transitions {
		for event, to := range events {
			sortedTransitions = append(sortedTransitions, transition{from, to, event})
		}
	}
	sort.Slice(sortedTransitions, func(i, j int) bool {
		if sortedTransitions[i].from != sortedTransitions[j].from {
			return sortedTransitions[i].from < sortedTransitions[j].from
		}
		if sortedTransitions[i].to != sortedTransitions[j].to {
			return sortedTransitions[i].to < sortedTransitions[j].to
		}
		return sortedTransitions[i].event < sortedTransitions[j].event
	})

	for _, t := range sortedTransitions {
		sb.WriteString(fmt.Sprintf("    \"%v\" -> \"%v\" [label=\"%v\"];\n", t.from, t.to, t.event))
	}

	sb.WriteString("}")

	return sb.String()
}

// GenerateDiagram generates a diagram representation of the state machine in the specified format
func GenerateDiagram[S, E comparable](sm *StateMachine[S, E], format DiagramFormat, currentState S) (string, error) {
	var generator DiagramGenerator

	switch format {
	case MermaidFormat:
		generator = &MermaidGenerator{}
	case DOTFormat:
		generator = &DOTGenerator{}
	default:
		return "", fmt.Errorf("unsupported diagram format")
	}

	// Convert states and transitions to string maps
	stringStates := make(map[string]struct{})
	for state := range sm.states {
		stringStates[fmt.Sprintf("%v", state)] = struct{}{}
	}

	stringTransitions := make(map[string]map[string]string)
	for from, events := range sm.transitions {
		fromStr := fmt.Sprintf("%v", from)
		stringTransitions[fromStr] = make(map[string]string)
		for event, transition := range events {
			eventStr := fmt.Sprintf("%v", event)
			toStr := fmt.Sprintf("%v", transition.to)
			stringTransitions[fromStr][eventStr] = toStr
		}
	}

	currentStateStr := fmt.Sprintf("%v", currentState)
	return generator.Generate(stringStates, stringTransitions, currentStateStr), nil
}
