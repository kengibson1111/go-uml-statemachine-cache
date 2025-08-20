package models

import "fmt"

// Vertex represents a vertex in a state machine (base type for states and pseudostates)
type Vertex struct {
	ID   string `json:"id" validate:"required"`
	Name string `json:"name" validate:"required"`
	Type string `json:"type" validate:"required"` // "state", "pseudostate", "finalstate"
	// Container *Region `json:"-"` // Parent region (not serialized)
}

// Validate validates the Vertex data integrity
func (v *Vertex) Validate() error {
	if v.ID == "" {
		return fmt.Errorf("Vertex ID cannot be empty")
	}
	if v.Name == "" {
		return fmt.Errorf("Vertex Name cannot be empty")
	}
	if v.Type == "" {
		return fmt.Errorf("Vertex Type cannot be empty")
	}

	// Validate type is one of the allowed values
	validTypes := map[string]bool{
		"state":       true,
		"pseudostate": true,
		"finalstate":  true,
	}
	if !validTypes[v.Type] {
		return fmt.Errorf("invalid Vertex Type: %s, must be one of: state, pseudostate, finalstate", v.Type)
	}

	return nil
}

// State represents a state in a state machine
type State struct {
	Vertex                                        // Embedded vertex
	IsComposite       bool                        `json:"is_composite"`
	IsOrthogonal      bool                        `json:"is_orthogonal"`
	IsSimple          bool                        `json:"is_simple"`
	IsSubmachineState bool                        `json:"is_submachine_state"`
	Regions           []*Region                   `json:"regions,omitempty"`
	Entry             *Behavior                   `json:"entry,omitempty"`
	Exit              *Behavior                   `json:"exit,omitempty"`
	DoActivity        *Behavior                   `json:"do_activity,omitempty"`
	Submachine        *StateMachine               `json:"submachine,omitempty"`
	Connections       []*ConnectionPointReference `json:"connections,omitempty"`
}

// Validate validates the State data integrity
func (s *State) Validate() error {
	// Validate embedded vertex
	if err := s.Vertex.Validate(); err != nil {
		return fmt.Errorf("invalid vertex in state: %w", err)
	}

	// Validate that type is "state"
	if s.Type != "state" {
		return fmt.Errorf("State must have type 'state', got: %s", s.Type)
	}

	// Validate regions if composite
	if s.IsComposite {
		for i, region := range s.Regions {
			if err := region.Validate(); err != nil {
				return fmt.Errorf("invalid region at index %d in composite state: %w", i, err)
			}
		}
	}

	// Validate behaviors
	if s.Entry != nil {
		if err := s.Entry.Validate(); err != nil {
			return fmt.Errorf("invalid entry behavior: %w", err)
		}
	}
	if s.Exit != nil {
		if err := s.Exit.Validate(); err != nil {
			return fmt.Errorf("invalid exit behavior: %w", err)
		}
	}
	if s.DoActivity != nil {
		if err := s.DoActivity.Validate(); err != nil {
			return fmt.Errorf("invalid do activity behavior: %w", err)
		}
	}

	// Validate submachine if present
	if s.Submachine != nil {
		if err := s.Submachine.Validate(); err != nil {
			return fmt.Errorf("invalid submachine: %w", err)
		}
	}

	// Validate connections
	for i, conn := range s.Connections {
		if err := conn.Validate(); err != nil {
			return fmt.Errorf("invalid connection at index %d: %w", i, err)
		}
	}

	return nil
}

// PseudostateKind represents the kind of pseudostate
type PseudostateKind string

const (
	PseudostateKindInitial        PseudostateKind = "initial"
	PseudostateKindDeepHistory    PseudostateKind = "deepHistory"
	PseudostateKindShallowHistory PseudostateKind = "shallowHistory"
	PseudostateKindJoin           PseudostateKind = "join"
	PseudostateKindFork           PseudostateKind = "fork"
	PseudostateKindJunction       PseudostateKind = "junction"
	PseudostateKindChoice         PseudostateKind = "choice"
	PseudostateKindEntryPoint     PseudostateKind = "entryPoint"
	PseudostateKindExitPoint      PseudostateKind = "exitPoint"
	PseudostateKindTerminate      PseudostateKind = "terminate"
)

// IsValid checks if the PseudostateKind is valid
func (pk PseudostateKind) IsValid() bool {
	validKinds := map[PseudostateKind]bool{
		PseudostateKindInitial:        true,
		PseudostateKindDeepHistory:    true,
		PseudostateKindShallowHistory: true,
		PseudostateKindJoin:           true,
		PseudostateKindFork:           true,
		PseudostateKindJunction:       true,
		PseudostateKindChoice:         true,
		PseudostateKindEntryPoint:     true,
		PseudostateKindExitPoint:      true,
		PseudostateKindTerminate:      true,
	}
	return validKinds[pk]
}

// Pseudostate represents a pseudostate in a state machine
type Pseudostate struct {
	Vertex                 // Embedded vertex
	Kind   PseudostateKind `json:"kind" validate:"required"`
}

// Validate validates the Pseudostate data integrity
func (ps *Pseudostate) Validate() error {
	// Validate embedded vertex
	if err := ps.Vertex.Validate(); err != nil {
		return fmt.Errorf("invalid vertex in pseudostate: %w", err)
	}

	// Validate that type is "pseudostate"
	if ps.Type != "pseudostate" {
		return fmt.Errorf("Pseudostate must have type 'pseudostate', got: %s", ps.Type)
	}

	// Validate kind
	if !ps.Kind.IsValid() {
		return fmt.Errorf("invalid PseudostateKind: %s", ps.Kind)
	}

	return nil
}

// FinalState represents a final state in a state machine
type FinalState struct {
	Vertex // Embedded vertex
}

// Validate validates the FinalState data integrity
func (fs *FinalState) Validate() error {
	// Validate embedded vertex
	if err := fs.Vertex.Validate(); err != nil {
		return fmt.Errorf("invalid vertex in final state: %w", err)
	}

	// Validate that type is "finalstate"
	if fs.Type != "finalstate" {
		return fmt.Errorf("FinalState must have type 'finalstate', got: %s", fs.Type)
	}

	return nil
}

// ConnectionPointReference represents a connection point reference
type ConnectionPointReference struct {
	Vertex                // Embedded vertex
	Entry  []*Pseudostate `json:"entry,omitempty"`
	Exit   []*Pseudostate `json:"exit,omitempty"`
	// State  *State         `json:"-"` // Parent state (not serialized)
}

// Validate validates the ConnectionPointReference data integrity
func (cpr *ConnectionPointReference) Validate() error {
	// Validate embedded vertex
	if err := cpr.Vertex.Validate(); err != nil {
		return fmt.Errorf("invalid vertex in connection point reference: %w", err)
	}

	// Validate entry pseudostates
	for i, entry := range cpr.Entry {
		if err := entry.Validate(); err != nil {
			return fmt.Errorf("invalid entry pseudostate at index %d: %w", i, err)
		}
	}

	// Validate exit pseudostates
	for i, exit := range cpr.Exit {
		if err := exit.Validate(); err != nil {
			return fmt.Errorf("invalid exit pseudostate at index %d: %w", i, err)
		}
	}

	return nil
}
