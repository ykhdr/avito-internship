package database

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"zadanie-6105/config"
	"zadanie-6105/model"
)

type postgresConnector struct {
	pool *pgxpool.Pool
}

func (c *postgresConnector) GetEmployeeByUsername(ctx context.Context, username string) (*model.Employee, error) {
	//TODO implement me
	panic("implement me")
}

func (c *postgresConnector) GetOrganizationById(ctx context.Context, id int) (*model.Organization, error) {
	//TODO implement me
	panic("implement me")
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
		return false, err
	}
	defer rows.Close()
	return rows.Next(), nil
}

func (c *postgresConnector) IsEmployeeExists(ctx context.Context, username string) (bool, error) {
	query := `SELECT EXISTS (SELECT 1 FROM employee WHERE username = $1)`
	rows, err := c.pool.Query(ctx, query, username)
	if err != nil {
		return false, err
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
