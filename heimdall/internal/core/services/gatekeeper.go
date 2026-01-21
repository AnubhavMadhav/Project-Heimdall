package services

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/AnubhavMadhav/project-heimdall/internal/adapters/postgres"
	"github.com/AnubhavMadhav/project-heimdall/internal/adapters/security"
)

type GatekeeperService struct {
	repo      *postgres.PostgresRepo
	validator *security.Validator
	logger    *slog.Logger
}

// NewGatekeeperService injects dependencies.
// Note: We accept the concrete types here but could accept interfaces for stricter decoupling.
func NewGatekeeperService(repo *postgres.PostgresRepo, validator *security.Validator, logger *slog.Logger) *GatekeeperService {
	return &GatekeeperService{
		repo:      repo,
		validator: validator,
		logger:    logger,
	}
}

func (s *GatekeeperService) ExecuteSafeQuery(ctx context.Context, query string) ([]map[string]interface{}, error) {
	// 1. SECURITY CHECK: Parse and Validate AST
	if err := s.validator.ValidateQuery(query); err != nil {
		s.logger.Warn("Blocked unsafe query attempt", "query", query, "reason", err)
		return nil, fmt.Errorf("security violation: %w", err)
	}

	// 2. EXECUTION: Run strictly on DB
	s.logger.Info("Executing safe query", "query", query)
	return s.repo.Execute(ctx, query)
}

func (s *GatekeeperService) ListTables(ctx context.Context) ([]string, error) {
	return s.repo.ListTables(ctx)
}

func (s *GatekeeperService) GetSchema(ctx context.Context, tableName string) (string, error) {
	return s.repo.GetSchema(ctx, tableName)
}
