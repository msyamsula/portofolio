package service

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/msyamsula/portofolio/backend-app/domain/agent/dto"
	"github.com/msyamsula/portofolio/backend-app/domain/agent/repository"
	postgresInfra "github.com/msyamsula/portofolio/backend-app/infrastructure/database/postgres"
	llmservice "github.com/msyamsula/portofolio/backend-app/infrastructure/llm/service"
	"github.com/msyamsula/portofolio/backend-app/infrastructure/repl/readline"
)

// agentService implements the Service interface
type agentService struct {
	repo         repository.Repository
	llm          llmservice.Service
	rl           *readline.Readline
	tx           *sqlx.Tx
	schema       string
	schemaLoaded bool
}

// New creates a new agent service
func New(repo repository.Repository) Service {
	return &agentService{
		repo: repo,
	}
}

// Connect establishes a database connection
func (s *agentService) Connect(ctx context.Context, cfg dto.ConnectionConfig) (*sqlx.DB, error) {
	var postgresCfg postgresInfra.Config

	if cfg.URI != "" {
		postgresCfg = postgresInfra.Config{
			URI: cfg.URI,
		}
	} else {
		postgresCfg = postgresInfra.Config{
			Host:     cfg.Host,
			Port:     cfg.Port,
			Database: cfg.Database,
			User:     cfg.User,
			Password: cfg.Password,
			SSLMode:  cfg.SSLMode,
		}
	}

	db, err := postgresInfra.NewClient(ctx, postgresCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Test connection
	if err := postgresInfra.TestConnection(ctx, db); err != nil {
		db.Close()
		return nil, fmt.Errorf("connection test failed: %w", err)
	}

	// Create new repository with this connection
	s.repo = repository.NewPostgresRepository(db)

	return db, nil
}

// ProcessInput processes user input and returns the response
func (s *agentService) ProcessInput(ctx context.Context, input string) (dto.QueryResponse, error) {
	return dto.QueryResponse{}, nil
}
