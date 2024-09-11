package database

import (
	"context"
	"zadanie-6105/model"
)

const (
	maxConns = 10
)

type Connector interface {
	GetEmployeeByUsername(ctx context.Context, username string) (*model.Employee, error)
	GetOrganizationById(ctx context.Context, id int) (*model.Organization, error)
	IsEmployeeInOrganization(ctx context.Context, username, organizationID string) (bool, error)
	IsEmployeeExists(ctx context.Context, username string) (bool, error)
}
