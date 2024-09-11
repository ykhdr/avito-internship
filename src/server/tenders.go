package server

import (
	"context"
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/samber/lo"
	"log/slog"
	"net/http"
	"slices"
	"strconv"
	"zadanie-6105/model"
)

var availableServiceTypes = []string{"Construction", "Delivery", "Manufacture"}

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
		if !slices.Contains(availableServiceTypes, st) {
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
	isCreatorExists, err := s.db.IsEmployeeExists(context.Background(), req.CreatorUsername)
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
	isInOrganizatoin, err := s.db.IsEmployeeInOrganization(context.Background(), req.CreatorUsername, req.OrganizationId)
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
	if !slices.Contains(availableServiceTypes, req.ServiceType) {
		w.WriteHeader(http.StatusBadRequest)
		resp := ErrResponse{Reason: "service type is not available"}
		_ = json.NewEncoder(w).Encode(resp)
		return
	}
	tender := requestToTender(&req)
	tender.Status = model.Created
	s.tenderStore.Save(tender)
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
	username := r.URL.Query().Get("username")
	if username == "" {
		w.WriteHeader(http.StatusUnauthorized)
		resp := ErrResponse{Reason: "username is empty"}
		_ = json.NewEncoder(w).Encode(resp)
		return
	}
	isEmployeeExists, err := s.db.IsEmployeeExists(context.Background(), username)
	if err != nil {
		slog.Warn("error checking employee exists", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		resp := ErrResponse{Reason: "error checking employee exists"}
		_ = json.NewEncoder(w).Encode(resp)
		return
	}
	if !isEmployeeExists {
		w.WriteHeader(http.StatusUnauthorized)
		resp := ErrResponse{Reason: "employee does not exist"}
		_ = json.NewEncoder(w).Encode(resp)
		return
	}
	employee, err := s.db.GetEmployeeByUsername(context.Background(), username)
	if err != nil {
		slog.Warn("error getting employee by username", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		resp := ErrResponse{Reason: "error getting employee by username"}
		_ = json.NewEncoder(w).Encode(resp)
		return
	}
	tenders := s.tenderStore.GetByCreatorID(limit, offset, employee.ID)
	resp := tendersToResponse(tenders)
	_ = json.NewEncoder(w).Encode(resp)
	w.WriteHeader(http.StatusOK)
}

func (s *Server) tenderStatus(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	if username == "" {
		w.WriteHeader(http.StatusUnauthorized)
		resp := ErrResponse{Reason: "username is empty"}
		_ = json.NewEncoder(w).Encode(resp)
		return
	}
	isEmployeeExists, err := s.db.IsEmployeeExists(context.Background(), username)
	if err != nil {
		slog.Warn("error checking employee exists", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		resp := ErrResponse{Reason: "error checking employee exists"}
		_ = json.NewEncoder(w).Encode(resp)
		return
	}
	if !isEmployeeExists {
		w.WriteHeader(http.StatusUnauthorized)
		resp := ErrResponse{Reason: "employee does not exist"}
		_ = json.NewEncoder(w).Encode(resp)
		return
	}
	employee, err := s.db.GetEmployeeByUsername(context.Background(), username)
	if err != nil {
		slog.Warn("error getting employee by username", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		resp := ErrResponse{Reason: "error getting employee by username"}
		_ = json.NewEncoder(w).Encode(resp)
		return
	}
	vars := mux.Vars(r)
	tenderID := vars["tenderId"]
	tender := s.tenderStore.GetByID(tenderID)
	if tender == nil {
		w.WriteHeader(http.StatusNotFound)
		resp := ErrResponse{Reason: "tender not found"}
		_ = json.NewEncoder(w).Encode(resp)
		return
	}
	if tender.CreatorID != employee.ID {
		w.WriteHeader(http.StatusForbidden)
		resp := ErrResponse{Reason: "employee is not creator of tender"}
		_ = json.NewEncoder(w).Encode(resp)
		return
	}
	resp := tenderToResponse(tender)
	_ = json.NewEncoder(w).Encode(resp)
	w.WriteHeader(http.StatusOK)
}

func (s *Server) updateTenderStatus(w http.ResponseWriter, r *http.Request) {

}

func tendersToResponse(tenders []*model.Tender) []*TenderResponse {
	return lo.Map(tenders, func(tender *model.Tender, _ int) *TenderResponse {
		return tenderToResponse(tender)
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
