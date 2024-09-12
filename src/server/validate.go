package server

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"log/slog"
	"net/http"
	"slices"
	"strconv"
	"zadanie-6105/database"
	"zadanie-6105/model"
)

const (
	MaxUsernameLength = 50
)

var availableServiceTypes = []string{"Construction", "Delivery", "Manufacture"}
var availableStatuses = []model.TenderStatus{model.Created, model.Published, model.Closed}

func IsValidStatus(status string) bool {
	return slices.Contains(availableStatuses, model.TenderStatus(status))
}

func IsValidServiceType(serviceType string) bool {
	return slices.Contains(availableServiceTypes, serviceType)
}

func IsTenderAvailable(t *model.Tender, employee *model.Employee) bool {
	return t.Status == model.Published || t.CreatorID == employee.ID
}

type Validator struct {
	w              http.ResponseWriter
	r              *http.Request
	organizationDb database.OrganizationConnector
}

func NewValidator(w http.ResponseWriter, r *http.Request, organizationDb database.OrganizationConnector) *Validator {
	return &Validator{w: w, r: r, organizationDb: organizationDb}
}

func (v *Validator) ValidateUsername(username string) bool {
	if username == "" {
		v.w.WriteHeader(http.StatusUnauthorized)
		resp := ErrResponse{Reason: "username is empty"}
		_ = json.NewEncoder(v.w).Encode(resp)
		return false
	}
	if len(username) > MaxUsernameLength {
		v.w.WriteHeader(http.StatusBadRequest)
		resp := ErrResponse{Reason: "username is too long. Max length is " + strconv.Itoa(MaxUsernameLength)}
		_ = json.NewEncoder(v.w).Encode(resp)
		return false
	}
	isEmployeeExists, err := v.organizationDb.IsEmployeeExists(context.Background(), username)
	if err != nil {
		slog.Warn("error checking employee exists", "error", err)
		v.w.WriteHeader(http.StatusInternalServerError)
		resp := ErrResponse{Reason: "error checking employee exists"}
		_ = json.NewEncoder(v.w).Encode(resp)
		return false
	}
	if !isEmployeeExists {
		v.w.WriteHeader(http.StatusUnauthorized)
		resp := ErrResponse{Reason: "employee does not exist"}
		_ = json.NewEncoder(v.w).Encode(resp)
		return false
	}
	return true
}

func (v *Validator) ValidateUuid(uuidValue string) bool {
	_, err := uuid.Parse(uuidValue)
	if err != nil {
		v.w.WriteHeader(http.StatusBadRequest)
		resp := ErrResponse{Reason: "tender id is not valid"}
		_ = json.NewEncoder(v.w).Encode(resp)
		return false
	}
	return true
}

func (v *Validator) ValidateStatus(status string) bool {
	if !IsValidStatus(status) {
		v.w.WriteHeader(http.StatusBadRequest)
		resp := ErrResponse{Reason: "status is not valid"}
		_ = json.NewEncoder(v.w).Encode(resp)
		return false
	}
	return true
}

func (v *Validator) ValidatePagination(limit, offset string) (bool, int, int) {
	var (
		limitInt  = 5
		offsetInt = 0
		err       error
	)
	if limit != "" {
		limitInt, err = strconv.Atoi(limit)
		if err != nil {
			v.w.WriteHeader(http.StatusBadRequest)
			resp := ErrResponse{Reason: "limit is not a number"}
			_ = json.NewEncoder(v.w).Encode(resp)
			return false, 0, 0
		}
	}
	if offset != "" {
		offsetInt, err = strconv.Atoi(offset)
		if err != nil {
			v.w.WriteHeader(http.StatusBadRequest)
			resp := ErrResponse{Reason: "offset is not a number"}
			_ = json.NewEncoder(v.w).Encode(resp)
			return false, 0, 0
		}
	}
	return true, limitInt, offsetInt
}
