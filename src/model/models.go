package model

import (
	"time"
)

type OrganizationType string

const (
	IE  OrganizationType = "IE"
	LLC OrganizationType = "LLC"
	JSC OrganizationType = "JSC"
)

type TenderStatus string

const (
	Created   TenderStatus = "Created"
	Published TenderStatus = "Published"
	Closed    TenderStatus = "Closed"
)

type Organization struct {
	ID          string
	Name        string
	Description string
	Type        OrganizationType
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type OrganizationResponsible struct {
	ID             string
	OrganizationID string
	UserID         string
}

type Employee struct {
	ID        string
	Username  string
	FirstName string
	LastName  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Tender struct {
	ID             string
	Name           string
	Description    string
	Status         TenderStatus
	ServiceType    string
	Version        int
	OrganizationID string
	CreatorID      string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
