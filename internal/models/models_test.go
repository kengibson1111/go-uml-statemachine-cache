package models

import (
	"testing"
	"time"
)

func TestStateMachine_Validate(t *testing.T) {
	tests := []struct {
		name    string
		sm      StateMachine
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid state machine",
			sm: StateMachine{
				Name:    "TestMachine",
				Version: "1.0",
				States: map[string]*State{
					"start": {ID: "start", Name: "Start", Type: "initial"},
					"end":   {ID: "end", Name: "End", Type: "final"},
				},
				Transitions: []Transition{
					{From: "start", To: "end", Event: "finish"},
				},
				Entities:  map[string]string{"entity1": "/cache/key/entity1"},
				CreatedAt: time.Now(),
			},
			wantErr: false,
		},
		{
			name: "empty name",
			sm: StateMachine{
				Name:    "",
				Version: "1.0",
				States: map[string]*State{
					"start": {ID: "start", Name: "Start", Type: "initial"},
				},
			},
			wantErr: true,
			errMsg:  "state machine name cannot be empty",
		},
		{
			name: "empty version",
			sm: StateMachine{
				Name:    "TestMachine",
				Version: "",
				States: map[string]*State{
					"start": {ID: "start", Name: "Start", Type: "initial"},
				},
			},
			wantErr: true,
			errMsg:  "state machine version cannot be empty",
		},
		{
			name: "no states",
			sm: StateMachine{
				Name:    "TestMachine",
				Version: "1.0",
				States:  map[string]*State{},
			},
			wantErr: true,
			errMsg:  "state machine must have at least one state",
		},
		{
			name: "state ID mismatch",
			sm: StateMachine{
				Name:    "TestMachine",
				Version: "1.0",
				States: map[string]*State{
					"start": {ID: "wrong_id", Name: "Start", Type: "initial"},
				},
			},
			wantErr: true,
			errMsg:  "state ID wrong_id does not match map key start",
		},
		{
			name: "invalid transition - non-existent from state",
			sm: StateMachine{
				Name:    "TestMachine",
				Version: "1.0",
				States: map[string]*State{
					"start": {ID: "start", Name: "Start", Type: "initial"},
				},
				Transitions: []Transition{
					{From: "nonexistent", To: "start", Event: "test"},
				},
			},
			wantErr: true,
			errMsg:  "transition references non-existent from state: nonexistent",
		},
		{
			name: "invalid transition - non-existent to state",
			sm: StateMachine{
				Name:    "TestMachine",
				Version: "1.0",
				States: map[string]*State{
					"start": {ID: "start", Name: "Start", Type: "initial"},
				},
				Transitions: []Transition{
					{From: "start", To: "nonexistent", Event: "test"},
				},
			},
			wantErr: true,
			errMsg:  "transition references non-existent to state: nonexistent",
		},
		{
			name: "empty entity ID",
			sm: StateMachine{
				Name:    "TestMachine",
				Version: "1.0",
				States: map[string]*State{
					"start": {ID: "start", Name: "Start", Type: "initial"},
				},
				Entities: map[string]string{"": "/cache/key"},
			},
			wantErr: true,
			errMsg:  "entity ID cannot be empty",
		},
		{
			name: "empty cache key for entity",
			sm: StateMachine{
				Name:    "TestMachine",
				Version: "1.0",
				States: map[string]*State{
					"start": {ID: "start", Name: "Start", Type: "initial"},
				},
				Entities: map[string]string{"entity1": ""},
			},
			wantErr: true,
			errMsg:  "cache key for entity entity1 cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.sm.Validate()
			if tt.wantErr {
				if err == nil {
					t.Errorf("StateMachine.Validate() expected error but got none")
					return
				}
				if tt.errMsg != "" && err.Error() != tt.errMsg {
					t.Errorf("StateMachine.Validate() error = %v, want %v", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("StateMachine.Validate() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestState_Validate(t *testing.T) {
	tests := []struct {
		name    string
		state   State
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid initial state",
			state:   State{ID: "start", Name: "Start", Type: "initial"},
			wantErr: false,
		},
		{
			name:    "valid final state",
			state:   State{ID: "end", Name: "End", Type: "final"},
			wantErr: false,
		},
		{
			name:    "valid simple state",
			state:   State{ID: "process", Name: "Processing", Type: "simple"},
			wantErr: false,
		},
		{
			name:    "valid composite state",
			state:   State{ID: "comp", Name: "Composite", Type: "composite"},
			wantErr: false,
		},
		{
			name:    "valid choice state",
			state:   State{ID: "choice", Name: "Choice", Type: "choice"},
			wantErr: false,
		},
		{
			name:    "valid fork state",
			state:   State{ID: "fork", Name: "Fork", Type: "fork"},
			wantErr: false,
		},
		{
			name:    "valid join state",
			state:   State{ID: "join", Name: "Join", Type: "join"},
			wantErr: false,
		},
		{
			name:    "empty ID",
			state:   State{ID: "", Name: "Start", Type: "initial"},
			wantErr: true,
			errMsg:  "state ID cannot be empty",
		},
		{
			name:    "empty name",
			state:   State{ID: "start", Name: "", Type: "initial"},
			wantErr: true,
			errMsg:  "state name cannot be empty",
		},
		{
			name:    "empty type",
			state:   State{ID: "start", Name: "Start", Type: ""},
			wantErr: true,
			errMsg:  "state type cannot be empty",
		},
		{
			name:    "invalid type",
			state:   State{ID: "start", Name: "Start", Type: "invalid"},
			wantErr: true,
			errMsg:  "invalid state type: invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.state.Validate()
			if tt.wantErr {
				if err == nil {
					t.Errorf("State.Validate() expected error but got none")
					return
				}
				if tt.errMsg != "" && err.Error() != tt.errMsg {
					t.Errorf("State.Validate() error = %v, want %v", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("State.Validate() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestTransition_Validate(t *testing.T) {
	tests := []struct {
		name       string
		transition Transition
		wantErr    bool
		errMsg     string
	}{
		{
			name:       "valid transition",
			transition: Transition{From: "start", To: "end", Event: "finish"},
			wantErr:    false,
		},
		{
			name:       "valid self-transition",
			transition: Transition{From: "state1", To: "state1", Event: "self"},
			wantErr:    false,
		},
		{
			name:       "transition with condition",
			transition: Transition{From: "start", To: "end", Event: "finish", Condition: "x > 0"},
			wantErr:    false,
		},
		{
			name:       "transition with actions",
			transition: Transition{From: "start", To: "end", Event: "finish", Actions: []string{"log", "notify"}},
			wantErr:    false,
		},
		{
			name:       "empty from state",
			transition: Transition{From: "", To: "end", Event: "finish"},
			wantErr:    true,
			errMsg:     "transition from state cannot be empty",
		},
		{
			name:       "empty to state",
			transition: Transition{From: "start", To: "", Event: "finish"},
			wantErr:    true,
			errMsg:     "transition to state cannot be empty",
		},
		{
			name:       "empty event",
			transition: Transition{From: "start", To: "end", Event: ""},
			wantErr:    true,
			errMsg:     "transition event cannot be empty",
		},
		{
			name:       "invalid self-transition",
			transition: Transition{From: "state1", To: "state1", Event: "other"},
			wantErr:    true,
			errMsg:     "self-transition must have event 'self', got: other",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.transition.Validate()
			if tt.wantErr {
				if err == nil {
					t.Errorf("Transition.Validate() expected error but got none")
					return
				}
				if tt.errMsg != "" && err.Error() != tt.errMsg {
					t.Errorf("Transition.Validate() error = %v, want %v", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("Transition.Validate() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestStateMachine_ToJSON_FromJSON(t *testing.T) {
	original := StateMachine{
		Name:    "TestMachine",
		Version: "1.0",
		States: map[string]*State{
			"start": {ID: "start", Name: "Start", Type: "initial", Properties: map[string]interface{}{"color": "green"}},
			"end":   {ID: "end", Name: "End", Type: "final"},
		},
		Transitions: []Transition{
			{From: "start", To: "end", Event: "finish", Condition: "ready", Actions: []string{"cleanup"}},
		},
		Entities: map[string]string{
			"entity1": "/cache/key/entity1",
		},
		Metadata: map[string]interface{}{
			"author": "test",
			"tags":   []string{"test", "example"},
		},
		CreatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
	}

	// Test ToJSON
	jsonData, err := original.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON() error = %v", err)
	}

	// Test FromJSON
	var restored StateMachine
	err = restored.FromJSON(jsonData)
	if err != nil {
		t.Fatalf("FromJSON() error = %v", err)
	}

	// Verify the data was preserved
	if restored.Name != original.Name {
		t.Errorf("Name mismatch: got %v, want %v", restored.Name, original.Name)
	}
	if restored.Version != original.Version {
		t.Errorf("Version mismatch: got %v, want %v", restored.Version, original.Version)
	}
	if len(restored.States) != len(original.States) {
		t.Errorf("States count mismatch: got %v, want %v", len(restored.States), len(original.States))
	}
	if len(restored.Transitions) != len(original.Transitions) {
		t.Errorf("Transitions count mismatch: got %v, want %v", len(restored.Transitions), len(original.Transitions))
	}

	// Test invalid JSON
	var invalid StateMachine
	err = invalid.FromJSON([]byte("invalid json"))
	if err == nil {
		t.Error("FromJSON() expected error for invalid JSON but got none")
	}
}

func TestStateMachine_GetStateByName(t *testing.T) {
	sm := StateMachine{
		States: map[string]*State{
			"start": {ID: "start", Name: "Start State", Type: "initial"},
			"end":   {ID: "end", Name: "End State", Type: "final"},
		},
	}

	// Test exact match
	state, found := sm.GetStateByName("Start State")
	if !found {
		t.Error("GetStateByName() should find exact match")
	}
	if state.ID != "start" {
		t.Errorf("GetStateByName() returned wrong state: got %v, want start", state.ID)
	}

	// Test case-insensitive match
	state, found = sm.GetStateByName("start state")
	if !found {
		t.Error("GetStateByName() should find case-insensitive match")
	}
	if state.ID != "start" {
		t.Errorf("GetStateByName() returned wrong state: got %v, want start", state.ID)
	}

	// Test not found
	_, found = sm.GetStateByName("nonexistent")
	if found {
		t.Error("GetStateByName() should not find nonexistent state")
	}
}

func TestStateMachine_GetTransitions(t *testing.T) {
	sm := StateMachine{
		Transitions: []Transition{
			{From: "start", To: "middle", Event: "event1"},
			{From: "start", To: "end", Event: "event2"},
			{From: "middle", To: "end", Event: "event3"},
			{From: "end", To: "start", Event: "reset"},
		},
	}

	// Test GetTransitionsFromState
	fromStart := sm.GetTransitionsFromState("start")
	if len(fromStart) != 2 {
		t.Errorf("GetTransitionsFromState() got %v transitions, want 2", len(fromStart))
	}

	fromMiddle := sm.GetTransitionsFromState("middle")
	if len(fromMiddle) != 1 {
		t.Errorf("GetTransitionsFromState() got %v transitions, want 1", len(fromMiddle))
	}

	fromNonexistent := sm.GetTransitionsFromState("nonexistent")
	if len(fromNonexistent) != 0 {
		t.Errorf("GetTransitionsFromState() got %v transitions, want 0", len(fromNonexistent))
	}

	// Test GetTransitionsToState
	toEnd := sm.GetTransitionsToState("end")
	if len(toEnd) != 2 {
		t.Errorf("GetTransitionsToState() got %v transitions, want 2", len(toEnd))
	}

	toStart := sm.GetTransitionsToState("start")
	if len(toStart) != 1 {
		t.Errorf("GetTransitionsToState() got %v transitions, want 1", len(toStart))
	}

	toNonexistent := sm.GetTransitionsToState("nonexistent")
	if len(toNonexistent) != 0 {
		t.Errorf("GetTransitionsToState() got %v transitions, want 0", len(toNonexistent))
	}
}
