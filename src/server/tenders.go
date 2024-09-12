package server

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/gorilla/mux"
	"github.com/samber/lo"
	"log/slog"
	"net/http"
	"strconv"
	"zadanie-6105/model"
	"zadanie-6105/store"
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
	tenders := s.tenderStore.GetAll(limit, offset, serviceType)
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
	isCreatorExists, err := s.organizationDb.IsEmployeeExists(context.Background(), req.CreatorUsername)
	if err != nil {
		slog.Warn("error checking employee exists", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
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
	isInOrganizatoin, err := s.organizationDb.IsEmployeeInOrganization(context.Background(), req.CreatorUsername, req.OrganizationId)
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
	if !IsValidStatus(req.ServiceType) {
		w.WriteHeader(http.StatusBadRequest)
		resp := ErrResponse{Reason: "service type is not available"}
		_ = json.NewEncoder(w).Encode(resp)
		return
	}
	tender := requestToTender(&req)
	tender.Status = model.Created
	if err := s.tenderStore.Save(tender); err != nil {
		if errors.Is(store.TenderAlreadyExists, err) {
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
	validator := NewValidator(w, r, s.organizationDb)
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
	employee, err := s.organizationDb.GetEmployeeByUsername(context.Background(), username)
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
	tenders := s.tenderStore.GetByCreatorID(limit, offset, employee.ID)
	resp := tendersToResponse(tenders)
	_ = json.NewEncoder(w).Encode(resp)
	w.WriteHeader(http.StatusOK)
}

func (s *Server) tenderStatus(w http.ResponseWriter, r *http.Request) {
	validator := NewValidator(w, r, s.organizationDb)
	tenderId := mux.Vars(r)["tenderId"]
	username := r.URL.Query().Get("username")
	if !validator.ValidateUuid(tenderId) {
		return
	}
	if !validator.ValidateUsername(username) {
		return
	}
	employee, err := s.organizationDb.GetEmployeeByUsername(context.Background(), username)
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
	tender, err := s.tenderStore.GetByID(tenderId)
	if err != nil {
		if errors.Is(store.TenderNotFound, err) {
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
	validator := NewValidator(w, r, s.organizationDb)
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
	employee, err := s.organizationDb.GetEmployeeByUsername(context.Background(), username)
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
	tender, err := s.tenderStore.GetByID(tenderId)
	if err != nil {
		if errors.Is(store.TenderNotFound, err) {
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

	if err := s.tenderStore.Update(tender); err != nil {
		if errors.Is(store.TenderNotFound, err) {
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

func (s *Server) editTender(w http.ResponseWriter, r *http.Request) {
	validator := NewValidator(w, r, s.organizationDb)
	tenderId := mux.Vars(r)["tenderId"]
	username := r.URL.Query().Get("username")
	if !validator.ValidateUuid(tenderId) {
		return
	}
	if !validator.ValidateUsername(username) {
		return
	}
	employee, err := s.organizationDb.GetEmployeeByUsername(context.Background(), username)
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
	tender, err := s.tenderStore.GetByID(tenderId)
	if err != nil {
		if errors.Is(store.TenderNotFound, err) {
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

	if err := s.tenderStore.Update(tender); err != nil {
		if errors.Is(store.TenderNotFound, err) {
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
	validator := NewValidator(w, r, s.organizationDb)
	tenderId := mux.Vars(r)["tenderId"]
	username := r.URL.Query().Get("username")
	if !validator.ValidateUuid(tenderId) {
		return
	}
	if !validator.ValidateUsername(username) {
		return
	}
	employee, err := s.organizationDb.GetEmployeeByUsername(context.Background(), username)
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
	tender, err := s.tenderStore.GetByID(tenderId)
	if err != nil {
		if errors.Is(store.TenderNotFound, err) {
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
	verS := mux.Vars(r)["version"]
	ver, err := strconv.Atoi(verS)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		resp := ErrResponse{Reason: "version is not integer"}
		_ = json.NewEncoder(w).Encode(resp)
		return
	}
	newTender, err := s.tenderStore.Rollback(tenderId, ver)
	if err != nil {
		if errors.Is(store.TenderNotFound, err) {
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

func requestToTender(req *TenderRequest) *model.Tender {
	return &model.Tender{
		Name:           req.Name,
		Description:    req.Description,
		ServiceType:    req.ServiceType,
		OrganizationID: req.OrganizationId,
		CreatorID:      req.CreatorUsername,
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
