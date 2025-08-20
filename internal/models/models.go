package models

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// StateMachine represents a parsed state machine with all its components
type StateMachine struct {
	Name        string                 `json:"name"`
	Version     string                 `json:"version"`
	States      map[string]*State      `json:"states"`
	Transitions []Transition           `json:"transitions"`
	Entities    map[string]string      `json:"entities"` // entityID -> cache key mapping
	Metadata    map[string]interface{} `json:"metadata"`
	CreatedAt   time.Time              `json:"created_at"`
}

// State represents a single state in the state machine
type State struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Type       string                 `json:"type"`
	Properties map[string]interface{} `json:"properties"`
}

// Transition represents a transition between states
type Transition struct {
	From      string   `json:"from"`
	To        string   `json:"to"`
	Event     string   `json:"event"`
	Condition string   `json:"condition,omitempty"`
	Actions   []string `json:"actions,omitempty"`
}

// Validate validates the StateMachine data integrity
func (sm *StateMachine) Validate() error {
	if sm.Name == "" {
		return fmt.Errorf("state machine name cannot be empty")
	}

	if sm.Version == "" {
		return fmt.Errorf("state machine version cannot be empty")
	}

	if len(sm.States) == 0 {
		return fmt.Errorf("state machine must have at least one state")
	}

	// Validate all states
	for stateID, state := range sm.States {
		if err := state.Validate(); err != nil {
			return fmt.Errorf("invalid state %s: %w", stateID, err)
		}

		// Ensure state ID matches the map key
		if state.ID != stateID {
			return fmt.Errorf("state ID %s does not match map key %s", state.ID, stateID)
		}
	}

	// Validate all transitions
	for i, transition := range sm.Transitions {
		if err := transition.Validate(); err != nil {
			return fmt.Errorf("invalid transition at index %d: %w", i, err)
		}

		// Ensure transition references valid states
		if _, exists := sm.States[transition.From]; !exists {
			return fmt.Errorf("transition references non-existent from state: %s", transition.From)
		}
		if _, exists := sm.States[transition.To]; !exists {
			return fmt.Errorf("transition references non-existent to state: %s", transition.To)
		}
	}

	// Validate entities mapping
	if sm.Entities != nil {
		for entityID, cacheKey := range sm.Entities {
			if entityID == "" {
				return fmt.Errorf("entity ID cannot be empty")
			}
			if cacheKey == "" {
				return fmt.Errorf("cache key for entity %s cannot be empty", entityID)
			}
		}
	}

	return nil
}

// Validate validates the State data integrity
func (s *State) Validate() error {
	if s.ID == "" {
		return fmt.Errorf("state ID cannot be empty")
	}

	if s.Name == "" {
		return fmt.Errorf("state name cannot be empty")
	}

	if s.Type == "" {
		return fmt.Errorf("state type cannot be empty")
	}

	// Validate state type is one of the expected values
	validTypes := []string{"initial", "final", "simple", "composite", "choice", "fork", "join"}
	isValidType := false
	for _, validType := range validTypes {
		if s.Type == validType {
			isValidType = true
			break
		}
	}
	if !isValidType {
		return fmt.Errorf("invalid state type: %s", s.Type)
	}

	return nil
}

// Validate validates the Transition data integrity
func (t *Transition) Validate() error {
	if t.From == "" {
		return fmt.Errorf("transition from state cannot be empty")
	}

	if t.To == "" {
		return fmt.Errorf("transition to state cannot be empty")
	}

	if t.Event == "" {
		return fmt.Errorf("transition event cannot be empty")
	}

	// Validate that from and to are different (no self-transitions without explicit handling)
	if t.From == t.To && t.Event != "self" {
		return fmt.Errorf("self-transition must have event 'self', got: %s", t.Event)
	}

	return nil
}

// ToJSON serializes the StateMachine to JSON
func (sm *StateMachine) ToJSON() ([]byte, error) {
	return json.Marshal(sm)
}

// FromJSON deserializes JSON data into a StateMachine
func (sm *StateMachine) FromJSON(data []byte) error {
	if err := json.Unmarshal(data, sm); err != nil {
		return fmt.Errorf("failed to unmarshal state machine: %w", err)
	}

	// Validate after unmarshaling
	if err := sm.Validate(); err != nil {
		return fmt.Errorf("validation failed after unmarshaling: %w", err)
	}

	return nil
}

// GetStateByName returns a state by its name (case-insensitive search)
func (sm *StateMachine) GetStateByName(name string) (*State, bool) {
	for _, state := range sm.States {
		if strings.EqualFold(state.Name, name) {
			return state, true
		}
	}
	return nil, false
}

// GetTransitionsFromState returns all transitions originating from a specific state
func (sm *StateMachine) GetTransitionsFromState(stateID string) []Transition {
	var transitions []Transition
	for _, transition := range sm.Transitions {
		if transition.From == stateID {
			transitions = append(transitions, transition)
		}
	}
	return transitions
}

// GetTransitionsToState returns all transitions leading to a specific state
func (sm *StateMachine) GetTransitionsToState(stateID string) []Transition {
	var transitions []Transition
	for _, transition := range sm.Transitions {
		if transition.To == stateID {
			transitions = append(transitions, transition)
		}
	}
	return transitions
}
