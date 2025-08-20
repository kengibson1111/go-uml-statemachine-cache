package models

import "testing"

func TestVertex_Validate(t *testing.T) {
	tests := []struct {
		name    string
		vertex  *Vertex
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid vertex",
			vertex: &Vertex{
				ID:   "v1",
				Name: "TestVertex",
				Type: "state",
			},
			wantErr: false,
		},
		{
			name: "empty ID",
			vertex: &Vertex{
				Name: "TestVertex",
				Type: "state",
			},
			wantErr: true,
			errMsg:  "Vertex ID cannot be empty",
		},
		{
			name: "empty Name",
			vertex: &Vertex{
				ID:   "v1",
				Type: "state",
			},
			wantErr: true,
			errMsg:  "Vertex Name cannot be empty",
		},
		{
			name: "empty Type",
			vertex: &Vertex{
				ID:   "v1",
				Name: "TestVertex",
			},
			wantErr: true,
			errMsg:  "Vertex Type cannot be empty",
		},
		{
			name: "invalid Type",
			vertex: &Vertex{
				ID:   "v1",
				Name: "TestVertex",
				Type: "invalid",
			},
			wantErr: true,
			errMsg:  "invalid Vertex Type: invalid, must be one of: state, pseudostate, finalstate",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.vertex.Validate()
			if tt.wantErr {
				if err == nil {
					t.Errorf("Vertex.Validate() expected error but got none")
					return
				}
				if tt.errMsg != "" && err.Error() != tt.errMsg {
					t.Errorf("Vertex.Validate() error = %v, want %v", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("Vertex.Validate() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestState_Validate(t *testing.T) {
	tests := []struct {
		name    string
		state   *State
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid simple state",
			state: &State{
				Vertex: Vertex{
					ID:   "s1",
					Name: "TestState",
					Type: "state",
				},
				IsSimple: true,
			},
			wantErr: false,
		},
		{
			name: "valid composite state with regions",
			state: &State{
				Vertex: Vertex{
					ID:   "s1",
					Name: "TestState",
					Type: "state",
				},
				IsComposite: true,
				Regions: []*Region{
					{
						ID:   "r1",
						Name: "TestRegion",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid vertex type",
			state: &State{
				Vertex: Vertex{
					ID:   "s1",
					Name: "TestState",
					Type: "pseudostate",
				},
			},
			wantErr: true,
			errMsg:  "State must have type 'state', got: pseudostate",
		},
		{
			name: "invalid entry behavior",
			state: &State{
				Vertex: Vertex{
					ID:   "s1",
					Name: "TestState",
					Type: "state",
				},
				Entry: &Behavior{
					// Missing required fields
				},
			},
			wantErr: true,
			errMsg:  "invalid entry behavior: Behavior ID cannot be empty",
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

func TestPseudostateKind_IsValid(t *testing.T) {
	tests := []struct {
		name string
		pk   PseudostateKind
		want bool
	}{
		{"initial", PseudostateKindInitial, true},
		{"deepHistory", PseudostateKindDeepHistory, true},
		{"shallowHistory", PseudostateKindShallowHistory, true},
		{"join", PseudostateKindJoin, true},
		{"fork", PseudostateKindFork, true},
		{"junction", PseudostateKindJunction, true},
		{"choice", PseudostateKindChoice, true},
		{"entryPoint", PseudostateKindEntryPoint, true},
		{"exitPoint", PseudostateKindExitPoint, true},
		{"terminate", PseudostateKindTerminate, true},
		{"invalid", PseudostateKind("invalid"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.pk.IsValid(); got != tt.want {
				t.Errorf("PseudostateKind.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPseudostate_Validate(t *testing.T) {
	tests := []struct {
		name    string
		ps      *Pseudostate
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid pseudostate",
			ps: &Pseudostate{
				Vertex: Vertex{
					ID:   "ps1",
					Name: "TestPseudostate",
					Type: "pseudostate",
				},
				Kind: PseudostateKindInitial,
			},
			wantErr: false,
		},
		{
			name: "invalid vertex type",
			ps: &Pseudostate{
				Vertex: Vertex{
					ID:   "ps1",
					Name: "TestPseudostate",
					Type: "state",
				},
				Kind: PseudostateKindInitial,
			},
			wantErr: true,
			errMsg:  "Pseudostate must have type 'pseudostate', got: state",
		},
		{
			name: "invalid kind",
			ps: &Pseudostate{
				Vertex: Vertex{
					ID:   "ps1",
					Name: "TestPseudostate",
					Type: "pseudostate",
				},
				Kind: PseudostateKind("invalid"),
			},
			wantErr: true,
			errMsg:  "invalid PseudostateKind: invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.ps.Validate()
			if tt.wantErr {
				if err == nil {
					t.Errorf("Pseudostate.Validate() expected error but got none")
					return
				}
				if tt.errMsg != "" && err.Error() != tt.errMsg {
					t.Errorf("Pseudostate.Validate() error = %v, want %v", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("Pseudostate.Validate() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestFinalState_Validate(t *testing.T) {
	tests := []struct {
		name    string
		fs      *FinalState
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid final state",
			fs: &FinalState{
				Vertex: Vertex{
					ID:   "fs1",
					Name: "TestFinalState",
					Type: "finalstate",
				},
			},
			wantErr: false,
		},
		{
			name: "invalid vertex type",
			fs: &FinalState{
				Vertex: Vertex{
					ID:   "fs1",
					Name: "TestFinalState",
					Type: "state",
				},
			},
			wantErr: true,
			errMsg:  "FinalState must have type 'finalstate', got: state",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fs.Validate()
			if tt.wantErr {
				if err == nil {
					t.Errorf("FinalState.Validate() expected error but got none")
					return
				}
				if tt.errMsg != "" && err.Error() != tt.errMsg {
					t.Errorf("FinalState.Validate() error = %v, want %v", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("FinalState.Validate() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestConnectionPointReference_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cpr     *ConnectionPointReference
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid connection point reference",
			cpr: &ConnectionPointReference{
				Vertex: Vertex{
					ID:   "cpr1",
					Name: "TestConnectionPoint",
					Type: "pseudostate",
				},
				Entry: []*Pseudostate{
					{
						Vertex: Vertex{
							ID:   "entry1",
							Name: "EntryPoint",
							Type: "pseudostate",
						},
						Kind: PseudostateKindEntryPoint,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid entry pseudostate",
			cpr: &ConnectionPointReference{
				Vertex: Vertex{
					ID:   "cpr1",
					Name: "TestConnectionPoint",
					Type: "pseudostate",
				},
				Entry: []*Pseudostate{
					{
						Vertex: Vertex{
							// Missing required fields
						},
					},
				},
			},
			wantErr: true,
			errMsg:  "invalid entry pseudostate at index 0: invalid vertex in pseudostate: Vertex ID cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cpr.Validate()
			if tt.wantErr {
				if err == nil {
					t.Errorf("ConnectionPointReference.Validate() expected error but got none")
					return
				}
				if tt.errMsg != "" && err.Error() != tt.errMsg {
					t.Errorf("ConnectionPointReference.Validate() error = %v, want %v", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("ConnectionPointReference.Validate() unexpected error = %v", err)
				}
			}
		})
	}
}
