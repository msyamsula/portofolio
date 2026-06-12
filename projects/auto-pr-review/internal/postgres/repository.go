package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"auto-pr-review/internal/config"
	"auto-pr-review/internal/service"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	createGitHubLinkStatement = "create_github_link"
	createGitHubLinkSQL       = `INSERT INTO github_links (url) VALUES ($1) RETURNING id, url`
	listGitHubLinksStatement  = "list_github_links"
	listGitHubLinksSQL        = `SELECT id, url FROM github_links ORDER BY id`
	deleteGitHubLinkStatement = "delete_github_link"
	deleteGitHubLinkSQL       = `DELETE FROM github_links WHERE id = $1 RETURNING id`
	uniqueViolationCode       = "23505"
)

type Repository struct {
	pool        *pgxpool.Pool
	execTimeout time.Duration
}

func NewRepository(ctx context.Context, cfg config.DatabaseConfig) (*Repository, error) {
	poolConfig, err := pgxpool.ParseConfig(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("parse postgres config: %w", err)
	}

	poolConfig.MaxConns = cfg.MaxConns
	poolConfig.MinConns = cfg.MinConns
	poolConfig.MaxConnLifetime = cfg.MaxConnLifetime
	poolConfig.MaxConnIdleTime = cfg.MaxConnIdleTime
	setupPool, err := pgxpool.NewWithConfig(ctx, poolConfig.Copy())
	if err != nil {
		return nil, fmt.Errorf("create postgres setup pool: %w", err)
	}

	setupCtx, cancel := context.WithTimeout(ctx, cfg.ExecTimeout)
	defer cancel()
	if err := setupPool.Ping(setupCtx); err != nil {
		setupPool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}
	if err := ensureSchema(setupCtx, setupPool); err != nil {
		setupPool.Close()
		return nil, err
	}
	setupPool.Close()

	poolConfig.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		if _, err := conn.Prepare(ctx, createGitHubLinkStatement, createGitHubLinkSQL); err != nil {
			return fmt.Errorf("prepare %s: %w", createGitHubLinkStatement, err)
		}
		if _, err := conn.Prepare(ctx, listGitHubLinksStatement, listGitHubLinksSQL); err != nil {
			return fmt.Errorf("prepare %s: %w", listGitHubLinksStatement, err)
		}
		if _, err := conn.Prepare(ctx, deleteGitHubLinkStatement, deleteGitHubLinkSQL); err != nil {
			return fmt.Errorf("prepare %s: %w", deleteGitHubLinkStatement, err)
		}
		return nil
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("create postgres pool: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, cfg.ExecTimeout)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres prepared pool: %w", err)
	}

	return &Repository{
		pool:        pool,
		execTimeout: cfg.ExecTimeout,
	}, nil
}

func (r *Repository) Close() {
	r.pool.Close()
}

func (r *Repository) CreateGitHubLink(ctx context.Context, githubURL string) (service.GitHubLink, error) {
	queryCtx, cancel := context.WithTimeout(ctx, r.execTimeout)
	defer cancel()

	var link service.GitHubLink
	err := r.pool.QueryRow(
		queryCtx,
		createGitHubLinkStatement,
		githubURL,
	).Scan(&link.ID, &link.URL)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == uniqueViolationCode {
			return service.GitHubLink{}, service.ErrGitHubLinkAlreadyExists
		}
		return service.GitHubLink{}, fmt.Errorf("insert github link: %w", err)
	}

	return link, nil
}

func (r *Repository) ListGitHubLinks(ctx context.Context) ([]service.GitHubLink, error) {
	queryCtx, cancel := context.WithTimeout(ctx, r.execTimeout)
	defer cancel()

	rows, err := r.pool.Query(queryCtx, listGitHubLinksStatement)
	if err != nil {
		return nil, fmt.Errorf("query github links: %w", err)
	}
	defer rows.Close()

	links := make([]service.GitHubLink, 0)
	for rows.Next() {
		var link service.GitHubLink
		if err := rows.Scan(&link.ID, &link.URL); err != nil {
			return nil, fmt.Errorf("scan github link: %w", err)
		}
		links = append(links, link)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate github links: %w", err)
	}

	return links, nil
}

func (r *Repository) DeleteGitHubLink(ctx context.Context, id int64) error {
	queryCtx, cancel := context.WithTimeout(ctx, r.execTimeout)
	defer cancel()

	var deletedID int64
	err := r.pool.QueryRow(
		queryCtx,
		deleteGitHubLinkStatement,
		id,
	).Scan(&deletedID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return service.ErrGitHubLinkNotFound
		}
		return fmt.Errorf("delete github link: %w", err)
	}

	return nil
}

func ensureSchema(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS github_links (
			id BIGSERIAL PRIMARY KEY,
			url TEXT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now()
		);
		CREATE UNIQUE INDEX IF NOT EXISTS github_links_url_unique_idx ON github_links (url);
	`)
	if err != nil {
		return fmt.Errorf("ensure postgres schema: %w", err)
	}

	return nil
}
