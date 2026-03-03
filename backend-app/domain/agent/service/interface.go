package service

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/msyamsula/portofolio/backend-app/domain/agent/dto"
)

// Service defines the interface for agent business logic
//
//go:generate mockgen -source=interface.go -destination=../../../mock/agent_service_mock.go -package=mock -mock_names Service=MockAgentService
type Service interface {
	// Connect establishes a database connection
	Connect(ctx context.Context, cfg dto.ConnectionConfig) (*sqlx.DB, error)
}
