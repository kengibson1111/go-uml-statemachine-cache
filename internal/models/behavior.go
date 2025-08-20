package models

import "fmt"

// Constraint represents a constraint (guard condition)
type Constraint struct {
	ID            string `json:"id" validate:"required"`
	Name          string `json:"name,omitempty"`
	Specification string `json:"specification" validate:"required"`
	Language      string `json:"language,omitempty"`
}

// Validate validates the Constraint data integrity
func (c *Constraint) Validate() error {
	if c.ID == "" {
		return fmt.Errorf("Constraint ID cannot be empty")
	}
	if c.Specification == "" {
		return fmt.Errorf("Constraint Specification cannot be empty")
	}
	return nil
}

// Behavior represents a behavior (action/activity)
type Behavior struct {
	ID            string `json:"id" validate:"required"`
	Name          string `json:"name,omitempty"`
	Specification string `json:"specification" validate:"required"`
	Language      string `json:"language,omitempty"`
}

// Validate validates the Behavior data integrity
func (b *Behavior) Validate() error {
	if b.ID == "" {
		return fmt.Errorf("Behavior ID cannot be empty")
	}
	if b.Specification == "" {
		return fmt.Errorf("Behavior Specification cannot be empty")
	}
	return nil
}

// Effect is an alias for Behavior to maintain semantic clarity
type Effect = Behavior
