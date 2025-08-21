package models

import (
	"testing"
	"time"
)

func TestStateMachine_Validate(t *testing.T) {
	tests := []struct {
		name    string
		sm      *StateMachine
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid state machine",
			sm: &StateMachine{
				ID:        "sm1",
				Name:      "TestStateMachine",
				Version:   "1.0",
				Regions:   []*Region{},
				Entities:  make(map[string]string),
				Metadata:  make(map[string]interface{}),
				CreatedAt: time.Now(),
			},
			wantErr: false,
		},
		{
			name: "empty ID",
			sm: &StateMachine{
				Name:    "TestStateMachine",
				Version: "1.0",
			},
			wantErr: true,
			errMsg:  "StateMachine ID cannot be empty",
		},
		{
			name: "empty Name",
			sm: &StateMachine{
				ID:      "sm1",
				Version: "1.0",
			},
			wantErr: true,
			errMsg:  "StateMachine Name cannot be empty",
		},
		{
			name: "empty Version",
			sm: &StateMachine{
				ID:   "sm1",
				Name: "TestStateMachine",
			},
			wantErr: true,
			errMsg:  "StateMachine Version cannot be empty",
		},
		{
			name: "invalid region",
			sm: &StateMachine{
				ID:      "sm1",
				Name:    "TestStateMachine",
				Version: "1.0",
				Regions: []*Region{
					{
						// Missing ID and Name
					},
				},
			},
			wantErr: true,
			errMsg:  "invalid region at index 0: Region ID cannot be empty",
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

func TestRegion_Validate(t *testing.T) {
	tests := []struct {
		name    string
		region  *Region
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid region",
			region: &Region{
				ID:          "r1",
				Name:        "TestRegion",
				States:      []*State{},
				Transitions: []*Transition{},
				Vertices:    []*Vertex{},
			},
			wantErr: false,
		},
		{
			name: "empty ID",
			region: &Region{
				Name: "TestRegion",
			},
			wantErr: true,
			errMsg:  "Region ID cannot be empty",
		},
		{
			name: "empty Name",
			region: &Region{
				ID: "r1",
			},
			wantErr: true,
			errMsg:  "Region Name cannot be empty",
		},
		{
			name: "invalid state",
			region: &Region{
				ID:   "r1",
				Name: "TestRegion",
				States: []*State{
					{
						Vertex: Vertex{
							// Missing required fields
						},
					},
				},
			},
			wantErr: true,
			errMsg:  "invalid state at index 0: invalid vertex in state: Vertex ID cannot be empty",
		},
		{
			name: "invalid transition",
			region: &Region{
				ID:   "r1",
				Name: "TestRegion",
				Transitions: []*Transition{
					{
						// Missing required fields
					},
				},
			},
			wantErr: true,
			errMsg:  "invalid transition at index 0: Transition ID cannot be empty",
		},
		{
			name: "invalid vertex",
			region: &Region{
				ID:   "r1",
				Name: "TestRegion",
				Vertices: []*Vertex{
					{
						// Missing required fields
					},
				},
			},
			wantErr: true,
			errMsg:  "invalid vertex at index 0: Vertex ID cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.region.Validate()
			if tt.wantErr {
				if err == nil {
					t.Errorf("Region.Validate() expected error but got none")
					return
				}
				if tt.errMsg != "" && err.Error() != tt.errMsg {
					t.Errorf("Region.Validate() error = %v, want %v", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("Region.Validate() unexpected error = %v", err)
				}
			}
		})
	}
}
