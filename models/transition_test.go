package models

import "testing"

func TestTransitionKind_IsValid(t *testing.T) {
	tests := []struct {
		name string
		tk   TransitionKind
		want bool
	}{
		{"internal", TransitionKindInternal, true},
		{"local", TransitionKindLocal, true},
		{"external", TransitionKindExternal, true},
		{"invalid", TransitionKind("invalid"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.tk.IsValid(); got != tt.want {
				t.Errorf("TransitionKind.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTransition_Validate(t *testing.T) {
	validSource := &Vertex{
		ID:   "source1",
		Name: "SourceState",
		Type: "state",
	}
	validTarget := &Vertex{
		ID:   "target1",
		Name: "TargetState",
		Type: "state",
	}

	tests := []struct {
		name       string
		transition *Transition
		wantErr    bool
		errMsg     string
	}{
		{
			name: "valid transition",
			transition: &Transition{
				ID:     "t1",
				Name:   "TestTransition",
				Source: validSource,
				Target: validTarget,
				Kind:   TransitionKindExternal,
			},
			wantErr: false,
		},
		{
			name: "empty ID",
			transition: &Transition{
				Name:   "TestTransition",
				Source: validSource,
				Target: validTarget,
				Kind:   TransitionKindExternal,
			},
			wantErr: true,
			errMsg:  "Transition ID cannot be empty",
		},
		{
			name: "nil source",
			transition: &Transition{
				ID:     "t1",
				Name:   "TestTransition",
				Target: validTarget,
				Kind:   TransitionKindExternal,
			},
			wantErr: true,
			errMsg:  "Transition Source cannot be nil",
		},
		{
			name: "nil target",
			transition: &Transition{
				ID:     "t1",
				Name:   "TestTransition",
				Source: validSource,
				Kind:   TransitionKindExternal,
			},
			wantErr: true,
			errMsg:  "Transition Target cannot be nil",
		},
		{
			name: "invalid source vertex",
			transition: &Transition{
				ID:     "t1",
				Name:   "TestTransition",
				Source: &Vertex{
					// Missing required fields
				},
				Target: validTarget,
				Kind:   TransitionKindExternal,
			},
			wantErr: true,
			errMsg:  "invalid source vertex: Vertex ID cannot be empty",
		},
		{
			name: "invalid target vertex",
			transition: &Transition{
				ID:     "t1",
				Name:   "TestTransition",
				Source: validSource,
				Target: &Vertex{
					// Missing required fields
				},
				Kind: TransitionKindExternal,
			},
			wantErr: true,
			errMsg:  "invalid target vertex: Vertex ID cannot be empty",
		},
		{
			name: "invalid kind",
			transition: &Transition{
				ID:     "t1",
				Name:   "TestTransition",
				Source: validSource,
				Target: validTarget,
				Kind:   TransitionKind("invalid"),
			},
			wantErr: true,
			errMsg:  "invalid TransitionKind: invalid",
		},
		{
			name: "invalid trigger",
			transition: &Transition{
				ID:     "t1",
				Name:   "TestTransition",
				Source: validSource,
				Target: validTarget,
				Kind:   TransitionKindExternal,
				Triggers: []*Trigger{
					{
						// Missing required fields
					},
				},
			},
			wantErr: true,
			errMsg:  "invalid trigger at index 0: Trigger ID cannot be empty",
		},
		{
			name: "invalid guard",
			transition: &Transition{
				ID:     "t1",
				Name:   "TestTransition",
				Source: validSource,
				Target: validTarget,
				Kind:   TransitionKindExternal,
				Guard:  &Constraint{
					// Missing required fields
				},
			},
			wantErr: true,
			errMsg:  "invalid guard constraint: Constraint ID cannot be empty",
		},
		{
			name: "invalid effect",
			transition: &Transition{
				ID:     "t1",
				Name:   "TestTransition",
				Source: validSource,
				Target: validTarget,
				Kind:   TransitionKindExternal,
				Effect: &Behavior{
					// Missing required fields
				},
			},
			wantErr: true,
			errMsg:  "invalid effect behavior: Behavior ID cannot be empty",
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
