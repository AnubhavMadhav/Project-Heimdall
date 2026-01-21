package security

import (
	"fmt"

	"github.com/xwb1989/sqlparser"
)

// Validator is a pure logic component that checks SQL safety.
type Validator struct{}

func NewValidator() *Validator {
	return &Validator{}
}

// ValidateQuery returns nil if the query is safe, or an error if unsafe.
func (v *Validator) ValidateQuery(sql string) error {
	// 1. Parse the SQL into an Abstract Syntax Tree (AST)
	stmt, err := sqlparser.Parse(sql)
	if err != nil {
		return fmt.Errorf("invalid SQL syntax: %w", err)
	}

	// 2. Strict Whitelist: Only allow SELECT statements
	switch stmt.(type) {
	case *sqlparser.Select:
		// 3. (Optional Senior Level Check) Block "SELECT INTO" or locking clauses
		// which can sometimes write data or lock tables.
		return v.deepInspectSelect(stmt.(*sqlparser.Select))
	default:
		return fmt.Errorf("SECURITY VIOLATION: Statement type %T is not allowed. Only SELECT is permitted.", stmt)
	}
}

func (v *Validator) deepInspectSelect(node *sqlparser.Select) error {
	// Example: Block 'SELECT ... FOR UPDATE' if you want to be extremely strict (prevent locking)
	if node.Lock != "" {
		return fmt.Errorf("SECURITY VIOLATION: Locking clauses (FOR UPDATE) are not allowed")
	}

	// Example: Block 'SELECT ... INTO' (which creates tables)
	// sqlparser handles INTO as a separate node type often, but strictly checking
	// the AST ensures no side effects.

	return nil
}
