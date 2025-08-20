package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestStateMachine_JSONSerialization(t *testing.T) {
	// Create a comprehensive state machine for testing JSON serialization
	now := time.Now()

	sm := &StateMachine{
		ID:      "sm1",
		Name:    "TestStateMachine",
		Version: "1.0.0",
		Regions: []*Region{
			{
				ID:   "region1",
				Name: "MainRegion",
				States: []*State{
					{
						Vertex: Vertex{
							ID:   "state1",
							Name: "InitialState",
							Type: "state",
						},
						IsSimple: true,
						Entry: &Behavior{
							ID:            "entry1",
							Name:          "EntryAction",
							Specification: "initialize()",
							Language:      "Java",
						},
					},
					{
						Vertex: Vertex{
							ID:   "state2",
							Name: "CompositeState",
							Type: "state",
						},
						IsComposite: true,
						Regions: []*Region{
							{
								ID:   "subregion1",
								Name: "SubRegion",
							},
						},
					},
				},
				Transitions: []*Transition{
					{
						ID:   "transition1",
						Name: "TestTransition",
						Source: &Vertex{
							ID:   "state1",
							Name: "InitialState",
							Type: "state",
						},
						Target: &Vertex{
							ID:   "state2",
							Name: "CompositeState",
							Type: "state",
						},
						Kind: TransitionKindExternal,
						Triggers: []*Trigger{
							{
								ID:   "trigger1",
								Name: "TestTrigger",
								Event: &Event{
									ID:   "event1",
									Name: "TestEvent",
									Type: EventTypeSignal,
								},
							},
						},
						Guard: &Constraint{
							ID:            "guard1",
							Name:          "TestGuard",
							Specification: "x > 0",
							Language:      "OCL",
						},
						Effect: &Behavior{
							ID:            "effect1",
							Name:          "TestEffect",
							Specification: "doAction()",
							Language:      "Java",
						},
					},
				},
				Vertices: []*Vertex{
					{
						ID:   "vertex1",
						Name: "TestVertex",
						Type: "pseudostate",
					},
				},
			},
		},
		Entities: map[string]string{
			"entity1": "/cache/path/entity1",
			"entity2": "/cache/path/entity2",
		},
		Metadata: map[string]interface{}{
			"author":      "test",
			"description": "Test state machine",
			"version":     1,
		},
		CreatedAt: now,
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(sm)
	if err != nil {
		t.Fatalf("Failed to marshal StateMachine to JSON: %v", err)
	}

	// Test JSON unmarshaling
	var unmarshaled StateMachine
	err = json.Unmarshal(jsonData, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON to StateMachine: %v", err)
	}

	// Validate the unmarshaled data
	if err := unmarshaled.Validate(); err != nil {
		t.Fatalf("Unmarshaled StateMachine failed validation: %v", err)
	}

	// Verify key fields
	if unmarshaled.ID != sm.ID {
		t.Errorf("ID mismatch: got %s, want %s", unmarshaled.ID, sm.ID)
	}
	if unmarshaled.Name != sm.Name {
		t.Errorf("Name mismatch: got %s, want %s", unmarshaled.Name, sm.Name)
	}
	if unmarshaled.Version != sm.Version {
		t.Errorf("Version mismatch: got %s, want %s", unmarshaled.Version, sm.Version)
	}
	if len(unmarshaled.Regions) != len(sm.Regions) {
		t.Errorf("Regions count mismatch: got %d, want %d", len(unmarshaled.Regions), len(sm.Regions))
	}
	if len(unmarshaled.Entities) != len(sm.Entities) {
		t.Errorf("Entities count mismatch: got %d, want %d", len(unmarshaled.Entities), len(sm.Entities))
	}

	// Verify nested structures
	if len(unmarshaled.Regions) > 0 {
		region := unmarshaled.Regions[0]
		if region.ID != "region1" {
			t.Errorf("Region ID mismatch: got %s, want region1", region.ID)
		}
		if len(region.States) != 2 {
			t.Errorf("States count mismatch: got %d, want 2", len(region.States))
		}
		if len(region.Transitions) != 1 {
			t.Errorf("Transitions count mismatch: got %d, want 1", len(region.Transitions))
		}
	}
}

func TestPseudostate_JSONSerialization(t *testing.T) {
	ps := &Pseudostate{
		Vertex: Vertex{
			ID:   "ps1",
			Name: "InitialPseudostate",
			Type: "pseudostate",
		},
		Kind: PseudostateKindInitial,
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(ps)
	if err != nil {
		t.Fatalf("Failed to marshal Pseudostate to JSON: %v", err)
	}

	// Test JSON unmarshaling
	var unmarshaled Pseudostate
	err = json.Unmarshal(jsonData, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON to Pseudostate: %v", err)
	}

	// Validate the unmarshaled data
	if err := unmarshaled.Validate(); err != nil {
		t.Fatalf("Unmarshaled Pseudostate failed validation: %v", err)
	}

	// Verify fields
	if unmarshaled.ID != ps.ID {
		t.Errorf("ID mismatch: got %s, want %s", unmarshaled.ID, ps.ID)
	}
	if unmarshaled.Kind != ps.Kind {
		t.Errorf("Kind mismatch: got %s, want %s", unmarshaled.Kind, ps.Kind)
	}
}

func TestFinalState_JSONSerialization(t *testing.T) {
	fs := &FinalState{
		Vertex: Vertex{
			ID:   "fs1",
			Name: "FinalState",
			Type: "finalstate",
		},
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(fs)
	if err != nil {
		t.Fatalf("Failed to marshal FinalState to JSON: %v", err)
	}

	// Test JSON unmarshaling
	var unmarshaled FinalState
	err = json.Unmarshal(jsonData, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON to FinalState: %v", err)
	}

	// Validate the unmarshaled data
	if err := unmarshaled.Validate(); err != nil {
		t.Fatalf("Unmarshaled FinalState failed validation: %v", err)
	}

	// Verify fields
	if unmarshaled.ID != fs.ID {
		t.Errorf("ID mismatch: got %s, want %s", unmarshaled.ID, fs.ID)
	}
}

func TestTransition_JSONSerialization(t *testing.T) {
	transition := &Transition{
		ID:   "t1",
		Name: "TestTransition",
		Source: &Vertex{
			ID:   "source",
			Name: "SourceState",
			Type: "state",
		},
		Target: &Vertex{
			ID:   "target",
			Name: "TargetState",
			Type: "state",
		},
		Kind: TransitionKindExternal,
		Triggers: []*Trigger{
			{
				ID:   "trigger1",
				Name: "TestTrigger",
				Event: &Event{
					ID:   "event1",
					Name: "TestEvent",
					Type: EventTypeSignal,
				},
			},
		},
		Guard: &Constraint{
			ID:            "guard1",
			Specification: "x > 0",
		},
		Effect: &Behavior{
			ID:            "effect1",
			Specification: "doAction()",
		},
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(transition)
	if err != nil {
		t.Fatalf("Failed to marshal Transition to JSON: %v", err)
	}

	// Test JSON unmarshaling
	var unmarshaled Transition
	err = json.Unmarshal(jsonData, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON to Transition: %v", err)
	}

	// Validate the unmarshaled data
	if err := unmarshaled.Validate(); err != nil {
		t.Fatalf("Unmarshaled Transition failed validation: %v", err)
	}

	// Verify fields
	if unmarshaled.ID != transition.ID {
		t.Errorf("ID mismatch: got %s, want %s", unmarshaled.ID, transition.ID)
	}
	if unmarshaled.Kind != transition.Kind {
		t.Errorf("Kind mismatch: got %s, want %s", unmarshaled.Kind, transition.Kind)
	}
	if len(unmarshaled.Triggers) != len(transition.Triggers) {
		t.Errorf("Triggers count mismatch: got %d, want %d", len(unmarshaled.Triggers), len(transition.Triggers))
	}
}
