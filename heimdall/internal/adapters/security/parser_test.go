package security

import (
	"testing"
)

func TestValidator_ValidateQuery(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name    string
		query   string
		wantErr bool
	}{
		// --- Happy Path (Should Pass) ---
		{
			name:    "Simple Select",
			query:   "SELECT * FROM users",
			wantErr: false,
		},
		{
			name:    "Select with Where",
			query:   "SELECT id, name FROM users WHERE id = 1",
			wantErr: false,
		},
		{
			name:    "Select Count",
			query:   "SELECT count(*) FROM orders",
			wantErr: false,
		},

		// --- Attack Vectors (Should Fail) ---
		{
			name:    "Delete Statement",
			query:   "DELETE FROM users WHERE id = 1",
			wantErr: true,
		},
		{
			name:    "Drop Table",
			query:   "DROP TABLE users",
			wantErr: true,
		},
		{
			name:    "Insert Data",
			query:   "INSERT INTO users (name) VALUES ('Hacker')",
			wantErr: true,
		},
		{
			name:    "SQL Injection Attempt (Stacked Query)",
			query:   "SELECT * FROM users; DROP TABLE users",
			wantErr: true, // Parser should fail on multiple statements or strictly catch the drop
		},
		{
			name:    "Update Statement",
			query:   "UPDATE users SET role = 'admin' WHERE name = 'Bob'",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateQuery(tt.query)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateQuery() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Run: go test ./internal/adapters/security/...
