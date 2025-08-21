package models

import "fmt"

// TransitionKind represents the kind of transition
type TransitionKind string

const (
	TransitionKindInternal TransitionKind = "internal"
	TransitionKindLocal    TransitionKind = "local"
	TransitionKindExternal TransitionKind = "external"
)

// IsValid checks if the TransitionKind is valid
func (tk TransitionKind) IsValid() bool {
	validKinds := map[TransitionKind]bool{
		TransitionKindInternal: true,
		TransitionKindLocal:    true,
		TransitionKindExternal: true,
	}
	return validKinds[tk]
}

// Transition represents a transition between vertices in a state machine
type Transition struct {
	ID       string         `json:"id" validate:"required"`
	Name     string         `json:"name,omitempty"`
	Source   *Vertex        `json:"source" validate:"required"`
	Target   *Vertex        `json:"target" validate:"required"`
	Kind     TransitionKind `json:"kind" validate:"required"`
	Triggers []*Trigger     `json:"triggers,omitempty"`
	Guard    *Constraint    `json:"guard,omitempty"`
	Effect   *Behavior      `json:"effect,omitempty"`
	// Container *Region       `json:"-"` // Parent region (not serialized)
}

// Validate validates the Transition data integrity
func (t *Transition) Validate() error {
	if t.ID == "" {
		return fmt.Errorf("Transition ID cannot be empty")
	}

	if t.Source == nil {
		return fmt.Errorf("Transition Source cannot be nil")
	}
	if err := t.Source.Validate(); err != nil {
		return fmt.Errorf("invalid source vertex: %w", err)
	}

	if t.Target == nil {
		return fmt.Errorf("Transition Target cannot be nil")
	}
	if err := t.Target.Validate(); err != nil {
		return fmt.Errorf("invalid target vertex: %w", err)
	}

	if !t.Kind.IsValid() {
		return fmt.Errorf("invalid TransitionKind: %s", t.Kind)
	}

	// Validate triggers
	for i, trigger := range t.Triggers {
		if err := trigger.Validate(); err != nil {
			return fmt.Errorf("invalid trigger at index %d: %w", i, err)
		}
	}

	// Validate guard constraint
	if t.Guard != nil {
		if err := t.Guard.Validate(); err != nil {
			return fmt.Errorf("invalid guard constraint: %w", err)
		}
	}

	// Validate effect behavior
	if t.Effect != nil {
		if err := t.Effect.Validate(); err != nil {
			return fmt.Errorf("invalid effect behavior: %w", err)
		}
	}

	return nil
}
