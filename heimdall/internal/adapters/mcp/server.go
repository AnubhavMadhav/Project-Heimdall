package mcp

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/AnubhavMadhav/project-heimdall/internal/core/ports"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type MCPServer struct {
	server     *server.MCPServer
	gatekeeper ports.Gatekeeper
	logger     *slog.Logger
}

func NewMCPServer(name, version string, gatekeeper ports.Gatekeeper, logger *slog.Logger) *MCPServer {
	s := server.NewMCPServer(name, version)
	return &MCPServer{
		server:     s,
		gatekeeper: gatekeeper,
		logger:     logger,
	}
}

// Start registers tools and begins listening on Stdio
func (s *MCPServer) Start() error {
	s.registerTools()
	return server.ServeStdio(s.server)
}

func (s *MCPServer) registerTools() {
	// --- Tool: safe_query ---
	s.server.AddTool(mcp.NewTool("safe_query",
		mcp.WithDescription("Executes a SELECT query after strict security checks. BLOCKS all other statements."),
		mcp.WithString("query", mcp.Required(), mcp.Description("The SQL SELECT statement")),
	), s.handleSafeQuery)

	// --- Tool: list_tables ---
	s.server.AddTool(mcp.NewTool("list_tables",
		mcp.WithDescription("Lists all public tables in the database."),
	), s.handleListTables)

	// --- Tool: get_schema ---
	s.server.AddTool(mcp.NewTool("get_schema",
		mcp.WithDescription("Returns the column schema for a specific table."),
		mcp.WithString("table_name", mcp.Required(), mcp.Description("The name of the table")),
	), s.handleGetSchema)
}

// Handlers are cleaner now because they are separated methods
func (s *MCPServer) handleSafeQuery(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, ok := request.Params.Arguments.(map[string]interface{})
	if !ok {
		return mcp.NewToolResultError("arguments must be a JSON object"), nil
	}

	query, ok := args["query"].(string)
	if !ok {
		return mcp.NewToolResultError("query argument must be a string"), nil
	}

	results, err := s.gatekeeper.ExecuteSafeQuery(ctx, query)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Query failed: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("%v", results)), nil
}

func (s *MCPServer) handleListTables(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	tables, err := s.gatekeeper.ListTables(ctx)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to list tables: %v", err)), nil
	}
	return mcp.NewToolResultText(fmt.Sprintf("Tables: %v", tables)), nil
}

func (s *MCPServer) handleGetSchema(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, ok := request.Params.Arguments.(map[string]interface{})
	if !ok {
		return mcp.NewToolResultError("arguments must be a JSON object"), nil
	}

	tableName, ok := args["table_name"].(string)
	if !ok {
		return mcp.NewToolResultError("table_name argument must be a string"), nil
	}

	schema, err := s.gatekeeper.GetSchema(ctx, tableName)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get schema: %v", err)), nil
	}
	return mcp.NewToolResultText(schema), nil
}
