package models

import "testing"

func TestConstraint_Validate(t *testing.T) {
	tests := []struct {
		name       string
		constraint *Constraint
		wantErr    bool
		errMsg     string
	}{
		{
			name: "valid constraint",
			constraint: &Constraint{
				ID:            "c1",
				Name:          "TestConstraint",
				Specification: "x > 0",
				Language:      "OCL",
			},
			wantErr: false,
		},
		{
			name: "valid constraint without name and language",
			constraint: &Constraint{
				ID:            "c1",
				Specification: "x > 0",
			},
			wantErr: false,
		},
		{
			name: "empty ID",
			constraint: &Constraint{
				Name:          "TestConstraint",
				Specification: "x > 0",
			},
			wantErr: true,
			errMsg:  "Constraint ID cannot be empty",
		},
		{
			name: "empty Specification",
			constraint: &Constraint{
				ID:   "c1",
				Name: "TestConstraint",
			},
			wantErr: true,
			errMsg:  "Constraint Specification cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.constraint.Validate()
			if tt.wantErr {
				if err == nil {
					t.Errorf("Constraint.Validate() expected error but got none")
					return
				}
				if tt.errMsg != "" && err.Error() != tt.errMsg {
					t.Errorf("Constraint.Validate() error = %v, want %v", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("Constraint.Validate() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestBehavior_Validate(t *testing.T) {
	tests := []struct {
		name     string
		behavior *Behavior
		wantErr  bool
		errMsg   string
	}{
		{
			name: "valid behavior",
			behavior: &Behavior{
				ID:            "b1",
				Name:          "TestBehavior",
				Specification: "doSomething()",
				Language:      "Java",
			},
			wantErr: false,
		},
		{
			name: "valid behavior without name and language",
			behavior: &Behavior{
				ID:            "b1",
				Specification: "doSomething()",
			},
			wantErr: false,
		},
		{
			name: "empty ID",
			behavior: &Behavior{
				Name:          "TestBehavior",
				Specification: "doSomething()",
			},
			wantErr: true,
			errMsg:  "Behavior ID cannot be empty",
		},
		{
			name: "empty Specification",
			behavior: &Behavior{
				ID:   "b1",
				Name: "TestBehavior",
			},
			wantErr: true,
			errMsg:  "Behavior Specification cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.behavior.Validate()
			if tt.wantErr {
				if err == nil {
					t.Errorf("Behavior.Validate() expected error but got none")
					return
				}
				if tt.errMsg != "" && err.Error() != tt.errMsg {
					t.Errorf("Behavior.Validate() error = %v, want %v", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("Behavior.Validate() unexpected error = %v", err)
				}
			}
		})
	}
}
