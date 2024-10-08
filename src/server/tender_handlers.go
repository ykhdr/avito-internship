package server

import (
	"encoding/json"
	"errors"
	"github.com/gorilla/mux"
	"github.com/samber/lo"
	"log/slog"
	"net/http"
	"strconv"
	"zadanie-6105/database"
	"zadanie-6105/model"
)

func (s *Server) tenders(w http.ResponseWriter, r *http.Request) {
	var (
		limit       = 5
		offset      = 0
		serviceType []string
		err         error
	)

	limitStr := r.URL.Query().Get("limit")
	if limitStr != "" {
		limit, err = strconv.Atoi(limitStr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			resp := ErrResponse{Reason: "limit is not integer"}
			_ = json.NewEncoder(w).Encode(resp)
			return
		}
	}
	offsetStr := r.URL.Query().Get("offset")
	if offsetStr != "" {
		offset, err = strconv.Atoi(offsetStr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			resp := ErrResponse{Reason: "offset is not integer"}
			_ = json.NewEncoder(w).Encode(resp)
			return
		}
	}
	serviceType = r.URL.Query()["service_type"]
	for _, st := range serviceType {
		if !IsValidServiceType(st) {
			w.WriteHeader(http.StatusBadRequest)
			resp := ErrResponse{Reason: "service type is not available"}
			_ = json.NewEncoder(w).Encode(resp)
			return
		}
	}
	if len(serviceType) == 0 {
		serviceType = availableServiceTypes
	}
	tenders, err := s.db.GetTenders(r.Context(), limit, offset, serviceType)
	if err != nil {
		slog.Warn("error getting tenders", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		resp := ErrResponse{Reason: "error getting tenders"}
		_ = json.NewEncoder(w).Encode(resp)
		return
	}
	resp := tendersToResponse(tenders)
	_ = json.NewEncoder(w).Encode(resp)
	w.WriteHeader(http.StatusOK)
}

func (s *Server) newTender(w http.ResponseWriter, r *http.Request) {
	var req TenderRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		slog.Warn("error decoding body", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		resp := ErrResponse{Reason: "error decoding body"}
		_ = json.NewEncoder(w).Encode(resp)
		return
	}
	isCreatorExists, err := s.db.IsEmployeeExists(r.Context(), req.CreatorUsername)
	if err != nil {
		slog.Warn("error checking employee exists", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		resp := ErrResponse{Reason: "error checking employee exists"}
		_ = json.NewEncoder(w).Encode(resp)
		return
	}
	if !isCreatorExists {
		w.WriteHeader(http.StatusUnauthorized)
		resp := ErrResponse{Reason: "employee does not exist"}
		_ = json.NewEncoder(w).Encode(resp)
		return
	}
	if !IsValidServiceType(req.ServiceType) {
		w.WriteHeader(http.StatusBadRequest)
		resp := ErrResponse{Reason: "service type is not available"}
		_ = json.NewEncoder(w).Encode(resp)
		return
	}
	employee, err := s.db.GetEmployeeByUsername(r.Context(), req.CreatorUsername)
	if err != nil {
		if errors.Is(err, database.ErrEmployeeNotFound) {
			w.WriteHeader(http.StatusNotFound)
			resp := ErrResponse{Reason: "employee not found"}
			_ = json.NewEncoder(w).Encode(resp)
			return
		}
		slog.Warn("error getting employee by username", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		resp := ErrResponse{Reason: "error getting employee by username"}
		_ = json.NewEncoder(w).Encode(resp)
		return
	}
	_, err = s.db.GetOrganizationById(r.Context(), req.OrganizationId)
	if err != nil {
		if errors.Is(err, database.ErrOrganizationNotFound) {
			w.WriteHeader(http.StatusBadRequest)
			resp := ErrResponse{Reason: "organization not found"}
			_ = json.NewEncoder(w).Encode(resp)
			return
		}
		slog.Warn("error getting organization by id", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		resp := ErrResponse{Reason: "error getting organization by id"}
		_ = json.NewEncoder(w).Encode(resp)
		return
	}
	isInOrganizatoin, err := s.db.IsEmployeeInOrganization(r.Context(), req.CreatorUsername, req.OrganizationId)
	if err != nil {
		slog.Warn("error checking employee in organization", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		resp := ErrResponse{Reason: "error checking employee in organization"}
		_ = json.NewEncoder(w).Encode(resp)
		return
	}
	if !isInOrganizatoin {
		w.WriteHeader(http.StatusForbidden)
		resp := ErrResponse{Reason: "employee is not in organization"}
		_ = json.NewEncoder(w).Encode(resp)
		return
	}
	tender := requestToTender(&req, employee.ID)
	tender.Status = model.TenderCreated
	if _, err := s.db.SaveTender(r.Context(), tender); err != nil {
		if errors.Is(err, database.ErrTenderAlreadyExists) {
			w.WriteHeader(http.StatusBadRequest)
			resp := ErrResponse{Reason: "tender already exists"}
			_ = json.NewEncoder(w).Encode(resp)
			return
		}
		slog.Warn("error saving tender", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		resp := ErrResponse{Reason: "error saving tender"}
		_ = json.NewEncoder(w).Encode(resp)
		return
	}
	resp := tenderToResponse(tender)
	_ = json.NewEncoder(w).Encode(resp)
	w.WriteHeader(http.StatusOK)
}

func (s *Server) myTenders(w http.ResponseWriter, r *http.Request) {
	var (
		limit  = 5
		offset = 0
		err    error
	)
	validator := NewValidator(w, r, s.db)
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	username := r.URL.Query().Get("username")
	var ok bool
	if ok, limit, offset = validator.ValidatePagination(limitStr, offsetStr); !ok {
		return
	}
	if !validator.ValidateUsername(username) {
		return
	}
	employee, err := s.db.GetEmployeeByUsername(r.Context(), username)
	if err != nil {
		if errors.Is(err, database.ErrEmployeeNotFound) {
			w.WriteHeader(http.StatusNotFound)
			resp := ErrResponse{Reason: "employee not found"}
			_ = json.NewEncoder(w).Encode(resp)
			return
		}
		slog.Warn("error getting employee by username", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		resp := ErrResponse{Reason: "error getting employee by username"}
		_ = json.NewEncoder(w).Encode(resp)
		return
	}
	tenders, err := s.db.GetTendersByCreatorID(r.Context(), limit, offset, employee.ID)
	resp := tendersToResponse(tenders)
	_ = json.NewEncoder(w).Encode(resp)
	w.WriteHeader(http.StatusOK)
}

func (s *Server) tenderStatus(w http.ResponseWriter, r *http.Request) {
	validator := NewValidator(w, r, s.db)
	tenderId := mux.Vars(r)["tenderId"]
	username := r.URL.Query().Get("username")
	if !validator.ValidateUuid(tenderId) {
		return
	}
	if !validator.ValidateUsername(username) {
		return
	}
	employee, err := s.db.GetEmployeeByUsername(r.Context(), username)
	if err != nil {
		slog.Warn("error getting employee by username", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		resp := ErrResponse{Reason: "error getting employee by username"}
		_ = json.NewEncoder(w).Encode(resp)
		return
	}
	if employee == nil {
		w.WriteHeader(http.StatusNotFound)
		resp := ErrResponse{Reason: "employee not found"}
		_ = json.NewEncoder(w).Encode(resp)
		return
	}
	tender, err := s.db.GetTenderByID(r.Context(), tenderId)
	if err != nil {
		if errors.Is(database.ErrTenderNotFound, err) {
			w.WriteHeader(http.StatusNotFound)
			resp := ErrResponse{Reason: "tender not found"}
			_ = json.NewEncoder(w).Encode(resp)
			return
		}
		slog.Warn("error getting tender by id", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		resp := ErrResponse{Reason: "error getting tender by id"}
		_ = json.NewEncoder(w).Encode(resp)
		return
	}
	if !IsTenderAvailable(tender, employee) {
		w.WriteHeader(http.StatusForbidden)
		resp := ErrResponse{Reason: "tender is not available"}
		_ = json.NewEncoder(w).Encode(resp)
		return
	}
	_ = json.NewEncoder(w).Encode(tender.Status)
	w.WriteHeader(http.StatusOK)
}

func (s *Server) updateTenderStatus(w http.ResponseWriter, r *http.Request) {
	validator := NewValidator(w, r, s.db)
	tenderId := mux.Vars(r)["tenderId"]
	username := r.URL.Query().Get("username")
	status := r.URL.Query().Get("status")
	if !validator.ValidateUuid(tenderId) {
		return
	}
	if !validator.ValidateUsername(username) {
		return
	}
	if !validator.ValidateStatus(status) {
		return
	}
	employee, err := s.db.GetEmployeeByUsername(r.Context(), username)
	if err != nil {
		slog.Warn("error getting employee by username", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		resp := ErrResponse{Reason: "error getting employee by username"}
		_ = json.NewEncoder(w).Encode(resp)
		return
	}

	if employee == nil {
		w.WriteHeader(http.StatusNotFound)
		resp := ErrResponse{Reason: "employee not found"}
		_ = json.NewEncoder(w).Encode(resp)
		return
	}
	tenderExists, err := s.db.IsTenderExists(r.Context(), tenderId)
	if err != nil {
		slog.Warn("error checking tender exists", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		resp := ErrResponse{Reason: "error checking tender exists"}
		_ = json.NewEncoder(w).Encode(resp)
		return
	}
	if !tenderExists {
		w.WriteHeader(http.StatusNotFound)
		resp := ErrResponse{Reason: "tender not found"}
		_ = json.NewEncoder(w).Encode(resp)
		return
	}
	tender, err := s.db.GetTenderByID(r.Context(), tenderId)
	if err != nil {
		if errors.Is(err, database.ErrTenderNotFound) {
			w.WriteHeader(http.StatusNotFound)
			resp := ErrResponse{Reason: "tender not found"}
			_ = json.NewEncoder(w).Encode(resp)
			return
		}
		slog.Warn("error getting tender by id", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		resp := ErrResponse{Reason: "error getting tender by id"}
		_ = json.NewEncoder(w).Encode(resp)
		return
	}
	if !IsTenderAvailable(tender, employee) {
		w.WriteHeader(http.StatusForbidden)
		resp := ErrResponse{Reason: "tender is not available"}
		_ = json.NewEncoder(w).Encode(resp)
		return
	}
	tender.Status = model.TenderStatus(status)
	if _, err := s.db.UpdateTender(r.Context(), tender); err != nil {
		if errors.Is(database.ErrTenderNotFound, err) {
			w.WriteHeader(http.StatusNotFound)
			resp := ErrResponse{Reason: "tender not found"}
			_ = json.NewEncoder(w).Encode(resp)
			return
		}
		slog.Warn("error updating tender", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		resp := ErrResponse{Reason: "error updating tender"}
		_ = json.NewEncoder(w).Encode(resp)
		return
	}

	resp := tenderToResponse(tender)
	_ = json.NewEncoder(w).Encode(resp)
	w.WriteHeader(http.StatusOK)
}

func (s *Server) editTender(w http.ResponseWriter, r *http.Request) {
	validator := NewValidator(w, r, s.db)
	tenderId := mux.Vars(r)["tenderId"]
	username := r.URL.Query().Get("username")
	if !validator.ValidateUuid(tenderId) {
		return
	}
	if !validator.ValidateUsername(username) {
		return
	}
	employee, err := s.db.GetEmployeeByUsername(r.Context(), username)
	if err != nil {
		slog.Warn("error getting employee by username", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		resp := ErrResponse{Reason: "error getting employee by username"}
		_ = json.NewEncoder(w).Encode(resp)
		return
	}
	if employee == nil {
		w.WriteHeader(http.StatusNotFound)
		resp := ErrResponse{Reason: "employee not found"}
		_ = json.NewEncoder(w).Encode(resp)
		return
	}
	tender, err := s.db.GetTenderByID(r.Context(), tenderId)
	if err != nil {
		if errors.Is(database.ErrTenderNotFound, err) {
			w.WriteHeader(http.StatusNotFound)
			resp := ErrResponse{Reason: "tender not found"}
			_ = json.NewEncoder(w).Encode(resp)
			return
		}
		slog.Warn("error getting tender by id", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		resp := ErrResponse{Reason: "error getting tender by id"}
		_ = json.NewEncoder(w).Encode(resp)
		return
	}
	isEmployeeInOrganization, err := s.db.IsEmployeeInOrganization(r.Context(), employee.Username, tender.OrganizationID)
	if err != nil {
		slog.Warn("error checking employee in organization", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		resp := ErrResponse{Reason: "error checking employee in organization"}
		_ = json.NewEncoder(w).Encode(resp)
		return
	}
	if !isEmployeeInOrganization {
		w.WriteHeader(http.StatusForbidden)
		resp := ErrResponse{Reason: "user not in organization. Tender is not available"}
		_ = json.NewEncoder(w).Encode(resp)
		return
	}
	var req TenderEditRequest
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		slog.Warn("error decoding body", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		resp := ErrResponse{Reason: "error decoding body"}
		_ = json.NewEncoder(w).Encode(resp)
		return
	}
	if req.ServiceType != "" && !IsValidServiceType(req.ServiceType) {
		w.WriteHeader(http.StatusBadRequest)
		resp := ErrResponse{Reason: "service type is not available"}
		_ = json.NewEncoder(w).Encode(resp)
		return
	}
	if req.Name != "" {
		if len(req.Name) > 100 {
			w.WriteHeader(http.StatusBadRequest)
			resp := ErrResponse{Reason: "name is too long. Max length is 100"}
			_ = json.NewEncoder(w).Encode(resp)
			return
		}
		tender.Name = req.Name
	}
	if req.Description != "" {
		tender.Description = req.Description
	}
	if req.ServiceType != "" {
		tender.ServiceType = req.ServiceType
	}

	if _, err := s.db.UpdateTender(r.Context(), tender); err != nil {
		if errors.Is(database.ErrTenderNotFound, err) {
			w.WriteHeader(http.StatusNotFound)
			resp := ErrResponse{Reason: "tender not found"}
			_ = json.NewEncoder(w).Encode(resp)
			return
		}
		slog.Warn("error updating employee", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		resp := ErrResponse{Reason: "error updating employee"}
		_ = json.NewEncoder(w).Encode(resp)
		return
	}

	resp := tenderToResponse(tender)
	_ = json.NewEncoder(w).Encode(resp)
	w.WriteHeader(http.StatusOK)
}

func (s *Server) rollbackVersion(w http.ResponseWriter, r *http.Request) {
	validator := NewValidator(w, r, s.db)
	tenderId := mux.Vars(r)["tenderId"]
	username := r.URL.Query().Get("username")
	if !validator.ValidateUuid(tenderId) {
		return
	}
	if !validator.ValidateUsername(username) {
		return
	}
	employee, err := s.db.GetEmployeeByUsername(r.Context(), username)
	if err != nil {
		slog.Warn("error getting employee by username", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		resp := ErrResponse{Reason: "error getting employee by username"}
		_ = json.NewEncoder(w).Encode(resp)
		return
	}
	if employee == nil {
		w.WriteHeader(http.StatusNotFound)
		resp := ErrResponse{Reason: "employee not found"}
		_ = json.NewEncoder(w).Encode(resp)
		return
	}
	tender, err := s.db.GetTenderByID(r.Context(), tenderId)
	if err != nil {
		if errors.Is(database.ErrTenderNotFound, err) {
			w.WriteHeader(http.StatusNotFound)
			resp := ErrResponse{Reason: "tender not found"}
			_ = json.NewEncoder(w).Encode(resp)
			return
		}
		slog.Warn("error getting tender by id", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		resp := ErrResponse{Reason: "error getting tender by id"}
		_ = json.NewEncoder(w).Encode(resp)
		return
	}
	isEmployeeInOrganization, err := s.db.IsEmployeeInOrganization(r.Context(), employee.Username, tender.OrganizationID)
	if err != nil {
		slog.Warn("error checking employee in organization", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		resp := ErrResponse{Reason: "error checking employee in organization"}
		_ = json.NewEncoder(w).Encode(resp)
		return
	}
	if !isEmployeeInOrganization {
		w.WriteHeader(http.StatusForbidden)
		resp := ErrResponse{Reason: "user not in organization. Tender is not available"}
		_ = json.NewEncoder(w).Encode(resp)
		return
	}
	verS := mux.Vars(r)["version"]
	ver, err := strconv.Atoi(verS)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		resp := ErrResponse{Reason: "version is not integer"}
		_ = json.NewEncoder(w).Encode(resp)
		return
	}
	newTender, err := s.db.RollbackTender(r.Context(), tenderId, ver)
	if err != nil {
		if errors.Is(database.ErrTenderNotFound, err) {
			w.WriteHeader(http.StatusNotFound)
			resp := ErrResponse{Reason: "tender not found"}
			_ = json.NewEncoder(w).Encode(resp)
			return
		}
		slog.Warn("error rolling back tender", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		resp := ErrResponse{Reason: "error rolling back tender"}
		_ = json.NewEncoder(w).Encode(resp)
		return
	}
	resp := tenderToResponse(newTender)
	_ = json.NewEncoder(w).Encode(resp)
	w.WriteHeader(http.StatusOK)
}

func tendersToResponse(tenders []model.Tender) []*TenderResponse {
	return lo.Map(tenders, func(tender model.Tender, _ int) *TenderResponse {
		return tenderToResponse(&tender)
	})
}

func requestToTender(req *TenderRequest, creatorId string) *model.Tender {
	return &model.Tender{
		Name:           req.Name,
		Description:    req.Description,
		ServiceType:    req.ServiceType,
		OrganizationID: req.OrganizationId,
		CreatorID:      creatorId,
	}
}

func tenderToResponse(tender *model.Tender) *TenderResponse {
	return &TenderResponse{
		ID:          tender.ID,
		Name:        tender.Name,
		Description: tender.Description,
		Status:      string(tender.Status),
		ServiceType: tender.ServiceType,
		Version:     tender.Version,
		CreatedAt:   JSONTime(tender.CreatedAt),
	}
}
