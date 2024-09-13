package database

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
	"log/slog"
	"zadanie-6105/config"
	"zadanie-6105/model"
)

var (
	ErrOrganizationNotFound = fmt.Errorf("organization not found")
	ErrTenderNotFound       = fmt.Errorf("tender not found")
	ErrTenderAlreadyExists  = fmt.Errorf("tender with same ID already exists")
)

type postgresConnector struct {
	DbConnector
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
	err = rows.Scan(&organization.ID, &organization.Name, &organization.Description, &organization.Type,
		&organization.CreatedAt, &organization.UpdatedAt)
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

func (c *postgresConnector) GetTenders(ctx context.Context, limit, offset int, serviceType []string) ([]model.Tender, error) {
	query := `
	SELECT id, name, description, service_type, status, organization_id, created_at, updated_at, creator_id
	FROM tender
	WHERE service_type = ANY($1)
	LIMIT $2 OFFSET $3
	`
	rows, err := c.pool.Query(ctx, query, serviceType, limit, offset)
	if err != nil {
		slog.Warn("error db query", "error", err, "query", query)
		return nil, errors.New("error db query")
	}
	defer rows.Close()
	var tenders []model.Tender
	for rows.Next() {
		var tender model.Tender
		err = rows.Scan(&tender.ID, &tender.Name, &tender.Description, &tender.ServiceType, &tender.Status,
			&tender.OrganizationID, &tender.CreatedAt, &tender.UpdatedAt, &tender.CreatorID)
		if err != nil {
			slog.Warn("error scan", "error", err)
			return nil, errors.New("error scan")
		}
		tenders = append(tenders, tender)
	}
	return tenders, nil
}

func (c *postgresConnector) GetMaxTenderVersion(ctx context.Context, id string) (int, error) {
	query := `SELECT MAX(version) FROM tender WHERE id = $1`
	rows, err := c.pool.Query(ctx, query, id)
	if err != nil {
		slog.Warn("error db query", "error", err, "query", query)
		return 0, errors.New("error db query")
	}
	defer rows.Close()
	if !rows.Next() {
		return 0, nil
	}
	var version int
	err = rows.Scan(&version)
	if err != nil {
		slog.Warn("error scan", "error", err)
		return 0, errors.New("error scan")
	}
	return version, nil
}

func (c *postgresConnector) GetTenderByID(ctx context.Context, id string) (*model.Tender, error) {
	maxVersion, err := c.GetMaxTenderVersion(ctx, id)
	if err != nil {
		return nil, err
	}
	query := `
	SELECT id, name, description, service_type, version,  status, organization_id, created_at, updated_at, creator_id
	FROM tender
	WHERE id = $1 AND version = $2
	`
	rows, err := c.pool.Query(ctx, query, id, maxVersion)
	if err != nil {
		slog.Warn("error db query", "error", err, "query", query)
		return nil, errors.New("error db query")
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, ErrTenderNotFound
	}
	var tender model.Tender
	err = rows.Scan(&tender.ID, &tender.Name, &tender.Description, &tender.ServiceType, &tender.Version, &tender.Status,
		&tender.OrganizationID, &tender.CreatedAt, &tender.UpdatedAt, &tender.CreatorID)
	if err != nil {
		slog.Warn("error scan", "error", err)
		return nil, errors.New("error scan")
	}
	return &tender, nil
}

func (c *postgresConnector) GetTendersByCreatorID(ctx context.Context, limit, offset int, creatorID string) ([]model.Tender, error) {
	query := `
	SELECT id, name, description, service_type,version,  status, organization_id, created_at, updated_at, creator_id
	FROM tender
	WHERE creator_id = $1 
	  AND version = (SELECT MAX(version) FROM tender as t WHERE t.id = tender.id) 
	LIMIT $2
	OFFSET $3
	`
	rows, err := c.pool.Query(ctx, query, creatorID, limit, offset)
	if err != nil {
		slog.Warn("error db query", "error", err, "query", query)
		return nil, errors.New("error db query")
	}
	defer rows.Close()
	var tenders []model.Tender
	for rows.Next() {
		var tender model.Tender
		err = rows.Scan(&tender.ID, &tender.Name, &tender.Description, &tender.ServiceType, &tender.Version, &tender.Status,
			&tender.OrganizationID, &tender.CreatedAt, &tender.UpdatedAt, &tender.CreatorID)
		if err != nil {
			slog.Warn("error scan", "error", err)
			return nil, errors.New("error scan")
		}
		tenders = append(tenders, tender)
	}
	return tenders, nil
}

func (c *postgresConnector) SaveTender(ctx context.Context, t *model.Tender) error {
	if _, err := c.GetTenderByID(ctx, t.ID); err == nil {
		return ErrTenderAlreadyExists
	}
	query := `
	INSERT INTO tender (name, description, service_type, status, organization_id, creator_id)
	VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := c.pool.Exec(ctx, query, t.Name, t.Description, t.ServiceType, t.Status, t.OrganizationID, t.CreatorID)
	if err != nil {
		slog.Warn("error db query", "error", err, "query", query)
		return errors.New("error db query")
	}
	return nil
}

func (c *postgresConnector) UpdateTender(ctx context.Context, t *model.Tender) error {
	tender, err := c.GetTenderByID(ctx, t.ID)
	if err != nil {
		return err
	}
	t.Version = tender.Version + 1
	return c.SaveTender(ctx, t)
}

func (c *postgresConnector) GetTenderByIdAndVersion(ctx context.Context, id string, version int) (*model.Tender, error) {
	query := `
	SELECT id, name, description, service_type, version,  status, organization_id, created_at, updated_at, creator_id
	FROM tender
	WHERE id = $1 AND version = $2
	`
	rows, err := c.pool.Query(ctx, query, id, version)
	if err != nil {
		slog.Warn("error db query", "error", err, "query", query)
		return nil, errors.New("error db query")
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, ErrTenderNotFound
	}
	var tender model.Tender
	err = rows.Scan(&tender.ID, &tender.Name, &tender.Description, &tender.ServiceType, &tender.Version, &tender.Status,
		&tender.OrganizationID, &tender.CreatedAt, &tender.UpdatedAt, &tender.CreatorID)
	if err != nil {
		slog.Warn("error scan", "error", err)
		return nil, errors.New("error scan")
	}
	return &tender, nil
}

func (c *postgresConnector) RollbackTender(ctx context.Context, id string, version int) (*model.Tender, error) {
	tender, err := c.GetTenderByIdAndVersion(ctx, id, version)
	if err != nil {
		return nil, err
	}
	maxVersion, err := c.GetMaxTenderVersion(ctx, id)
	if err != nil {
		return nil, err
	}
	tender.Version = maxVersion + 1
	return tender, c.SaveTender(ctx, tender)
}

func NewPostgresConnector(cfg *config.Config) (DbConnector, error) {
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
