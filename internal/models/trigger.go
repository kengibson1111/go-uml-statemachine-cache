package models

import "fmt"

// EventType represents the type of event
type EventType string

const (
	EventTypeCall       EventType = "call"
	EventTypeSignal     EventType = "signal"
	EventTypeChange     EventType = "change"
	EventTypeTime       EventType = "time"
	EventTypeAnyReceive EventType = "anyReceive"
)

// IsValid checks if the EventType is valid
func (et EventType) IsValid() bool {
	validTypes := map[EventType]bool{
		EventTypeCall:       true,
		EventTypeSignal:     true,
		EventTypeChange:     true,
		EventTypeTime:       true,
		EventTypeAnyReceive: true,
	}
	return validTypes[et]
}

// Event represents an event that can trigger a transition
type Event struct {
	ID   string    `json:"id" validate:"required"`
	Name string    `json:"name" validate:"required"`
	Type EventType `json:"type" validate:"required"`
}

// Validate validates the Event data integrity
func (e *Event) Validate() error {
	if e.ID == "" {
		return fmt.Errorf("Event ID cannot be empty")
	}
	if e.Name == "" {
		return fmt.Errorf("Event Name cannot be empty")
	}
	if !e.Type.IsValid() {
		return fmt.Errorf("invalid EventType: %s", e.Type)
	}
	return nil
}

// Trigger represents a trigger for a transition
type Trigger struct {
	ID    string `json:"id" validate:"required"`
	Name  string `json:"name" validate:"required"`
	Event *Event `json:"event" validate:"required"`
}

// Validate validates the Trigger data integrity
func (tr *Trigger) Validate() error {
	if tr.ID == "" {
		return fmt.Errorf("Trigger ID cannot be empty")
	}
	if tr.Name == "" {
		return fmt.Errorf("Trigger Name cannot be empty")
	}
	if tr.Event == nil {
		return fmt.Errorf("Trigger Event cannot be nil")
	}
	if err := tr.Event.Validate(); err != nil {
		return fmt.Errorf("invalid event: %w", err)
	}
	return nil
}
