package domain

import "fmt"

// SafeQueryResult wraps the database response to decouple the core from the specific DB driver format.
type SafeQueryResult struct {
	Rows     []map[string]interface{} `json:"rows"`
	RowCount int                      `json:"row_count"`
}

// NewSafeQueryResult creates a standardized result.
func NewSafeQueryResult(rows []map[string]interface{}) *SafeQueryResult {
	return &SafeQueryResult{
		Rows:     rows,
		RowCount: len(rows),
	}
}

// String returns a human-readable representation for the LLM.
func (r *SafeQueryResult) String() string {
	if r.RowCount == 0 {
		return "Query executed successfully. Result: 0 rows."
	}
	return fmt.Sprintf("Query executed successfully. Returned %d rows: %v", r.RowCount, r.Rows)
}
