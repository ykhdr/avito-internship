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
	TenderCreated   TenderStatus = "Created"
	TenderPublished TenderStatus = "Published"
	TenderClosed    TenderStatus = "Closed"
)

type BidStatus string

const (
	BidCreated   BidStatus = "Created"
	BidPublished BidStatus = "Published"
	BidCanceled  BidStatus = "Canceled"
)

type AuthorType string

const (
	AuthorUser         AuthorType = "User"
	AuthorOrganization AuthorType = "Organization"
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

type Bid struct {
	ID        string
	Name      string
	Status    BidStatus
	Author    AuthorType
	AuthorId  string
	TenderId  string
	Version   int
	CratedAt  time.Time
	UpdatedAt time.Time
}
