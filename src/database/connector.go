package database

import (
	"context"
	"zadanie-6105/model"
)

const (
	maxConns = 10
)

type DbConnector interface {
	GetEmployeeByUsername(ctx context.Context, username string) (*model.Employee, error)
	GetOrganizationById(ctx context.Context, id int) (*model.Organization, error)
	IsEmployeeInOrganization(ctx context.Context, username, organizationID string) (bool, error)
	IsEmployeeExists(ctx context.Context, username string) (bool, error)
	GetMaxTenderVersion(ctx context.Context, id string) (int, error)
	GetTenders(ctx context.Context, limit, offset int, serviceType []string) ([]model.Tender, error)
	GetTenderByID(ctx context.Context, id string) (*model.Tender, error)
	GetTendersByCreatorID(ctx context.Context, limit, offset int, creatorID string) ([]model.Tender, error)
	SaveTender(ctx context.Context, t *model.Tender) error
	UpdateTender(ctx context.Context, t *model.Tender) error
	GetTenderByIdAndVersion(ctx context.Context, id string, version int) (*model.Tender, error)
	RollbackTender(ctx context.Context, id string, version int) (*model.Tender, error)
}
