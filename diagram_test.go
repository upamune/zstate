package zstate_test

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/upamune/zstate"
)

var update = flag.Bool("update", false, "update golden files")

func TestGenerateDiagram(t *testing.T) {
	t.Parallel()

	sm := buildDoorStateMachine(t)

	tests := []struct {
		name         string
		format       zstate.DiagramFormat
		currentState DoorState
		goldenFile   string
	}{
		{
			name:         "Mermaid Diagram - Closed State",
			format:       zstate.MermaidFormat,
			currentState: Closed,
			goldenFile:   "testdata/mermaid_closed.golden",
		},
		{
			name:         "Mermaid Diagram - Open State",
			format:       zstate.MermaidFormat,
			currentState: Open,
			goldenFile:   "testdata/mermaid_open.golden",
		},
		{
			name:         "Mermaid Diagram - Locked State",
			format:       zstate.MermaidFormat,
			currentState: Locked,
			goldenFile:   "testdata/mermaid_locked.golden",
		},
		{
			name:         "DOT Diagram - Closed State",
			format:       zstate.DOTFormat,
			currentState: Closed,
			goldenFile:   "testdata/dot_closed.golden",
		},
		{
			name:         "DOT Diagram - Open State",
			format:       zstate.DOTFormat,
			currentState: Open,
			goldenFile:   "testdata/dot_open.golden",
		},
		{
			name:         "DOT Diagram - Locked State",
			format:       zstate.DOTFormat,
			currentState: Locked,
			goldenFile:   "testdata/dot_locked.golden",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diagram, err := zstate.GenerateDiagram(sm, tt.format, tt.currentState)
			if err != nil {
				t.Fatalf("Failed to generate diagram: %v", err)
			}

			if *update {
				err = os.MkdirAll(filepath.Dir(tt.goldenFile), 0755)
				if err != nil {
					t.Fatalf("Failed to create golden file directory: %v", err)
				}
				err = os.WriteFile(tt.goldenFile, []byte(diagram), 0644)
				if err != nil {
					t.Fatalf("Failed to update golden file: %v", err)
				}
			}

			expected, err := os.ReadFile(tt.goldenFile)
			if err != nil {
				t.Fatalf("Failed to read golden file: %v", err)
			}

			if diagram != string(expected) {
				t.Errorf("Generated diagram does not match golden file.\nExpected:\n%s\n\nGot:\n%s", expected, diagram)
			}
		})
	}
}

func TestGenerateDiagramErrors(t *testing.T) {
	t.Parallel()

	sm := buildDoorStateMachine(t)

	_, err := zstate.GenerateDiagram(sm, zstate.DiagramFormat(999), Closed)
	if err == nil {
		t.Fatal("Expected error for unsupported diagram format, got nil")
	}
	expectedErrMsg := "unsupported diagram format"
	if err.Error() != expectedErrMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedErrMsg, err.Error())
	}
}
