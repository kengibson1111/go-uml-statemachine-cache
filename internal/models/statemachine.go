package models

import (
	"fmt"
	"time"
)

// StateMachine represents a UML state machine
type StateMachine struct {
	ID        string                 `json:"id" validate:"required"`
	Name      string                 `json:"name" validate:"required"`
	Version   string                 `json:"version" validate:"required"`
	Regions   []*Region              `json:"regions"`
	Entities  map[string]string      `json:"entities"` // entityID -> cache key mapping
	Metadata  map[string]interface{} `json:"metadata"`
	CreatedAt time.Time              `json:"created_at"`
}

// Validate validates the StateMachine data integrity
func (sm *StateMachine) Validate() error {
	if sm.ID == "" {
		return fmt.Errorf("StateMachine ID cannot be empty")
	}
	if sm.Name == "" {
		return fmt.Errorf("StateMachine Name cannot be empty")
	}
	if sm.Version == "" {
		return fmt.Errorf("StateMachine Version cannot be empty")
	}

	// Validate regions
	for i, region := range sm.Regions {
		if err := region.Validate(); err != nil {
			return fmt.Errorf("invalid region at index %d: %w", i, err)
		}
	}

	return nil
}

// Region represents a region within a state machine
type Region struct {
	ID          string        `json:"id" validate:"required"`
	Name        string        `json:"name" validate:"required"`
	States      []*State      `json:"states"`
	Transitions []*Transition `json:"transitions"`
	Vertices    []*Vertex     `json:"vertices"`
}

// Validate validates the Region data integrity
func (r *Region) Validate() error {
	if r.ID == "" {
		return fmt.Errorf("Region ID cannot be empty")
	}
	if r.Name == "" {
		return fmt.Errorf("Region Name cannot be empty")
	}

	// Validate states
	for i, state := range r.States {
		if err := state.Validate(); err != nil {
			return fmt.Errorf("invalid state at index %d: %w", i, err)
		}
	}

	// Validate transitions
	for i, transition := range r.Transitions {
		if err := transition.Validate(); err != nil {
			return fmt.Errorf("invalid transition at index %d: %w", i, err)
		}
	}

	// Validate vertices
	for i, vertex := range r.Vertices {
		if err := vertex.Validate(); err != nil {
			return fmt.Errorf("invalid vertex at index %d: %w", i, err)
		}
	}

	return nil
}
