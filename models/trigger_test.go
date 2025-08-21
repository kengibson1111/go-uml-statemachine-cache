package models

import "testing"

func TestEventType_IsValid(t *testing.T) {
	tests := []struct {
		name string
		et   EventType
		want bool
	}{
		{"call", EventTypeCall, true},
		{"signal", EventTypeSignal, true},
		{"change", EventTypeChange, true},
		{"time", EventTypeTime, true},
		{"anyReceive", EventTypeAnyReceive, true},
		{"invalid", EventType("invalid"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.et.IsValid(); got != tt.want {
				t.Errorf("EventType.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEvent_Validate(t *testing.T) {
	tests := []struct {
		name    string
		event   *Event
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid event",
			event: &Event{
				ID:   "e1",
				Name: "TestEvent",
				Type: EventTypeSignal,
			},
			wantErr: false,
		},
		{
			name: "empty ID",
			event: &Event{
				Name: "TestEvent",
				Type: EventTypeSignal,
			},
			wantErr: true,
			errMsg:  "Event ID cannot be empty",
		},
		{
			name: "empty Name",
			event: &Event{
				ID:   "e1",
				Type: EventTypeSignal,
			},
			wantErr: true,
			errMsg:  "Event Name cannot be empty",
		},
		{
			name: "invalid Type",
			event: &Event{
				ID:   "e1",
				Name: "TestEvent",
				Type: EventType("invalid"),
			},
			wantErr: true,
			errMsg:  "invalid EventType: invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.event.Validate()
			if tt.wantErr {
				if err == nil {
					t.Errorf("Event.Validate() expected error but got none")
					return
				}
				if tt.errMsg != "" && err.Error() != tt.errMsg {
					t.Errorf("Event.Validate() error = %v, want %v", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("Event.Validate() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestTrigger_Validate(t *testing.T) {
	validEvent := &Event{
		ID:   "e1",
		Name: "TestEvent",
		Type: EventTypeSignal,
	}

	tests := []struct {
		name    string
		trigger *Trigger
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid trigger",
			trigger: &Trigger{
				ID:    "tr1",
				Name:  "TestTrigger",
				Event: validEvent,
			},
			wantErr: false,
		},
		{
			name: "empty ID",
			trigger: &Trigger{
				Name:  "TestTrigger",
				Event: validEvent,
			},
			wantErr: true,
			errMsg:  "Trigger ID cannot be empty",
		},
		{
			name: "empty Name",
			trigger: &Trigger{
				ID:    "tr1",
				Event: validEvent,
			},
			wantErr: true,
			errMsg:  "Trigger Name cannot be empty",
		},
		{
			name: "nil Event",
			trigger: &Trigger{
				ID:   "tr1",
				Name: "TestTrigger",
			},
			wantErr: true,
			errMsg:  "Trigger Event cannot be nil",
		},
		{
			name: "invalid Event",
			trigger: &Trigger{
				ID:    "tr1",
				Name:  "TestTrigger",
				Event: &Event{
					// Missing required fields
				},
			},
			wantErr: true,
			errMsg:  "invalid event: Event ID cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.trigger.Validate()
			if tt.wantErr {
				if err == nil {
					t.Errorf("Trigger.Validate() expected error but got none")
					return
				}
				if tt.errMsg != "" && err.Error() != tt.errMsg {
					t.Errorf("Trigger.Validate() error = %v, want %v", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("Trigger.Validate() unexpected error = %v", err)
				}
			}
		})
	}
}
