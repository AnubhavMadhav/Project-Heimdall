package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type PostgresRepo struct {
	conn *pgx.Conn
}

func NewPostgresRepo(ctx context.Context, connString string) (*PostgresRepo, error) {
	conn, err := pgx.Connect(ctx, connString)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}
	return &PostgresRepo{conn: conn}, nil
}

func (r *PostgresRepo) Close(ctx context.Context) error {
	return r.conn.Close(ctx)
}

func (r *PostgresRepo) Execute(ctx context.Context, query string) ([]map[string]interface{}, error) {
	rows, err := r.conn.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Convert rows to map[string]interface{} for dynamic JSON response
	fields := rows.FieldDescriptions()
	columns := make([]string, len(fields))
	for i, f := range fields {
		columns[i] = f.Name
	}

	var results []map[string]interface{}

	for rows.Next() {
		// Create a slice of interface{} to hold the values
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}

		rowMap := make(map[string]interface{})
		for i, col := range columns {
			var v interface{}
			val := values[i]

			// Handle Postgres bytes/text formatting if needed
			b, ok := val.([]byte)
			if ok {
				v = string(b)
			} else {
				v = val
			}
			rowMap[col] = v
		}
		results = append(results, rowMap)
	}

	return results, nil
}

// ListTables returns all public tables in the database.
func (r *PostgresRepo) ListTables(ctx context.Context) ([]string, error) {
	query := `
		SELECT table_name 
		FROM information_schema.tables 
		WHERE table_schema = 'public' 
		AND table_type = 'BASE TABLE';
	`
	rows, err := r.conn.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		tables = append(tables, name)
	}
	return tables, nil
}

// GetSchema returns the CREATE TABLE equivalent (or column list) for context.
func (r *PostgresRepo) GetSchema(ctx context.Context, tableName string) (string, error) {
	query := `
		SELECT column_name, data_type, is_nullable
		FROM information_schema.columns
		WHERE table_name = $1 AND table_schema = 'public'
		ORDER BY ordinal_position;
	`
	rows, err := r.conn.Query(ctx, query, tableName)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	var schemaBuilder string
	schemaBuilder += fmt.Sprintf("Table: %s\n", tableName)
	schemaBuilder += "Columns:\n"

	for rows.Next() {
		var col, dtype, nullable string
		if err := rows.Scan(&col, &dtype, &nullable); err != nil {
			return "", err
		}
		schemaBuilder += fmt.Sprintf(" - %s (%s, Nullable: %s)\n", col, dtype, nullable)
	}
	return schemaBuilder, nil
}
