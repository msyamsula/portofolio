package postgres

import (
	"context"
	"database/sql"
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
	createGitHubLinkSQL       = `INSERT INTO github_links (url) VALUES ($1) RETURNING id, url, periodic_interval`
	listGitHubLinksStatement  = "list_github_links"
	listGitHubLinksSQL        = `SELECT id, url, periodic_interval FROM github_links ORDER BY id`
	deleteGitHubLinkStatement = "delete_github_link"
	deleteGitHubLinkSQL       = `DELETE FROM github_links WHERE id = $1 RETURNING id`
	setPeriodicJobStatement   = "set_periodic_job"
	setPeriodicJobSQL         = `UPDATE github_links SET periodic_interval = $2 WHERE url = $1 RETURNING url, periodic_interval`
	clearPeriodicJobStatement = "clear_periodic_job"
	clearPeriodicJobSQL       = `UPDATE github_links SET periodic_interval = NULL WHERE url = $1 AND periodic_interval = $2 RETURNING url`
	listPeriodicJobsStatement = "list_periodic_jobs"
	listPeriodicJobsSQL       = `SELECT url, periodic_interval FROM github_links WHERE periodic_interval IS NOT NULL ORDER BY id`
	getPeriodicJobStatement   = "get_periodic_job"
	getPeriodicJobSQL         = `SELECT url, periodic_interval FROM github_links WHERE url = $1 AND periodic_interval IS NOT NULL`
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
		if _, err := conn.Prepare(ctx, setPeriodicJobStatement, setPeriodicJobSQL); err != nil {
			return fmt.Errorf("prepare %s: %w", setPeriodicJobStatement, err)
		}
		if _, err := conn.Prepare(ctx, clearPeriodicJobStatement, clearPeriodicJobSQL); err != nil {
			return fmt.Errorf("prepare %s: %w", clearPeriodicJobStatement, err)
		}
		if _, err := conn.Prepare(ctx, listPeriodicJobsStatement, listPeriodicJobsSQL); err != nil {
			return fmt.Errorf("prepare %s: %w", listPeriodicJobsStatement, err)
		}
		if _, err := conn.Prepare(ctx, getPeriodicJobStatement, getPeriodicJobSQL); err != nil {
			return fmt.Errorf("prepare %s: %w", getPeriodicJobStatement, err)
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
	var periodicInterval sql.NullString
	err := r.pool.QueryRow(
		queryCtx,
		createGitHubLinkStatement,
		githubURL,
	).Scan(&link.ID, &link.URL, &periodicInterval)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == uniqueViolationCode {
			return service.GitHubLink{}, service.ErrGitHubLinkAlreadyExists
		}
		return service.GitHubLink{}, fmt.Errorf("insert github link: %w", err)
	}
	link.PeriodicInterval = nullStringPtr(periodicInterval)

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
		var periodicInterval sql.NullString
		if err := rows.Scan(&link.ID, &link.URL, &periodicInterval); err != nil {
			return nil, fmt.Errorf("scan github link: %w", err)
		}
		link.PeriodicInterval = nullStringPtr(periodicInterval)
		links = append(links, link)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate github links: %w", err)
	}

	return links, nil
}

func (r *Repository) SetPeriodicJob(ctx context.Context, githubURL, period string) (service.PeriodicJob, error) {
	queryCtx, cancel := context.WithTimeout(ctx, r.execTimeout)
	defer cancel()

	var job service.PeriodicJob
	err := r.pool.QueryRow(
		queryCtx,
		setPeriodicJobStatement,
		githubURL,
		period,
	).Scan(&job.GitHubLink, &job.Period)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return service.PeriodicJob{}, service.ErrGitHubLinkNotFound
		}
		return service.PeriodicJob{}, fmt.Errorf("set periodic job: %w", err)
	}

	return job, nil
}

func (r *Repository) ClearPeriodicJob(ctx context.Context, githubURL, period string) error {
	queryCtx, cancel := context.WithTimeout(ctx, r.execTimeout)
	defer cancel()

	var clearedURL string
	err := r.pool.QueryRow(
		queryCtx,
		clearPeriodicJobStatement,
		githubURL,
		period,
	).Scan(&clearedURL)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		return fmt.Errorf("clear periodic job: %w", err)
	}

	return nil
}

func (r *Repository) ListPeriodicJobs(ctx context.Context) ([]service.PeriodicJob, error) {
	queryCtx, cancel := context.WithTimeout(ctx, r.execTimeout)
	defer cancel()

	rows, err := r.pool.Query(queryCtx, listPeriodicJobsStatement)
	if err != nil {
		return nil, fmt.Errorf("query periodic jobs: %w", err)
	}
	defer rows.Close()

	jobs := make([]service.PeriodicJob, 0)
	for rows.Next() {
		var job service.PeriodicJob
		if err := rows.Scan(&job.GitHubLink, &job.Period); err != nil {
			return nil, fmt.Errorf("scan periodic job: %w", err)
		}
		jobs = append(jobs, job)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate periodic jobs: %w", err)
	}

	return jobs, nil
}

func (r *Repository) GetPeriodicJob(ctx context.Context, githubURL string) (service.PeriodicJob, error) {
	queryCtx, cancel := context.WithTimeout(ctx, r.execTimeout)
	defer cancel()

	var job service.PeriodicJob
	err := r.pool.QueryRow(queryCtx, getPeriodicJobStatement, githubURL).Scan(&job.GitHubLink, &job.Period)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return service.PeriodicJob{}, service.ErrGitHubLinkNotFound
		}
		return service.PeriodicJob{}, fmt.Errorf("get periodic job: %w", err)
	}

	return job, nil
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
			periodic_interval TEXT,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now()
		);
		ALTER TABLE github_links ADD COLUMN IF NOT EXISTS periodic_interval TEXT;
		CREATE UNIQUE INDEX IF NOT EXISTS github_links_url_unique_idx ON github_links (url);
	`)
	if err != nil {
		return fmt.Errorf("ensure postgres schema: %w", err)
	}

	return nil
}

func nullStringPtr(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}

	result := value.String
	return &result
}
