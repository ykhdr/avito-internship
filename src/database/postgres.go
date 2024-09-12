package database

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
	"log/slog"
	"zadanie-6105/config"
	"zadanie-6105/model"
)

type postgresConnector struct {
	pool *pgxpool.Pool
}

func (c *postgresConnector) GetEmployeeByUsername(ctx context.Context, username string) (*model.Employee, error) {
	query := `
	SELECT id, username, first_name, last_name, created_at, updated_at
	FROM employee
	WHERE username = $1
	`
	rows, err := c.pool.Query(ctx, query, username)
	if err != nil {
		slog.Warn("error db query", "error", err, "query", query)
		return nil, errors.New("error db query")
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, nil
	}
	var employee model.Employee
	err = rows.Scan(&employee.ID, &employee.Username, &employee.FirstName, &employee.LastName, &employee.CreatedAt, &employee.UpdatedAt)
	if err != nil {
		slog.Warn("error scan", "error", err)
		return nil, errors.New("error scan")
	}
	return &employee, nil
}

func (c *postgresConnector) GetOrganizationById(ctx context.Context, id int) (*model.Organization, error) {
	query := `
	SELECT id, name, description, type, created_at, updated_at
	FROM organization
	WHERE id = $1
	`
	rows, err := c.pool.Query(ctx, query, id)
	if err != nil {
		slog.Warn("error db query", "error", err, "query", query)
		return nil, errors.New("error db query")
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, nil
	}
	var organization model.Organization
	err = rows.Scan(&organization.ID, &organization.Name, &organization.Description, &organization.Type, &organization.CreatedAt, &organization.UpdatedAt)
	if err != nil {
		slog.Warn("error scan", "error", err)
		return nil, errors.New("error scan")
	}
	return &organization, nil
}

func (c *postgresConnector) IsEmployeeInOrganization(ctx context.Context, username, organizationID string) (bool, error) {
	query := `
	SELECT EXISTS (SELECT 1
               FROM organization_responsible
               WHERE user_id = (SELECT id FROM employee WHERE employee.username = $1)
                 AND organization_id = $2)
	`
	rows, err := c.pool.Query(ctx, query, username, organizationID)
	if err != nil {
		slog.Warn("error db query", "error", err, "query", query)
		return false, errors.New("error db query")
	}
	defer rows.Close()
	return rows.Next(), nil
}

func (c *postgresConnector) IsEmployeeExists(ctx context.Context, username string) (bool, error) {
	query := `SELECT EXISTS (SELECT 1 FROM employee WHERE username = $1)`
	rows, err := c.pool.Query(ctx, query, username)
	if err != nil {
		slog.Warn("error db query", "error", err, "query", query)
		return false, errors.New("error db query")
	}
	defer rows.Close()
	return rows.Next(), nil
}

func NewPostgresConnector(cfg *config.Config) (OrganizationConnector, error) {
	pgxConfig, err := pgxpool.ParseConfig(cfg.PostgresConn)
	if err != nil {
		return nil, err
	}
	pgxConfig.MaxConns = maxConns
	pool, err := pgxpool.NewWithConfig(context.Background(), pgxConfig)
	if err != nil {
		return nil, err
	}
	return &postgresConnector{pool: pool}, nil
}
