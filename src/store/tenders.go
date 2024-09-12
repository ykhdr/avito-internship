package store

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/samber/lo"
	"log/slog"
	"slices"
	"sync"
	"time"
	"zadanie-6105/model"
)

const defaultVersion = 1

var (
	TenderNotFound      = fmt.Errorf("tender not found")
	TenderAlreadyExists = fmt.Errorf("tender with same ID already exists")
)

type tenderId struct {
	Id  string
	Ver int
}

type TenderStore struct {
	tenders map[tenderId]model.Tender
	mu      sync.RWMutex
}

func NewTendersStore() *TenderStore {
	return &TenderStore{}
}

func (s *TenderStore) GetAll(limit, offset int, serviceType []string) []model.Tender {
	var res []model.Tender
	if len(serviceType) != 0 {
		s.mu.RLock()
		for _, t := range s.tenders {
			if slices.Contains(serviceType, t.ServiceType) {
				res = append(res, t)
			}
		}
		s.mu.RUnlock()
	} else {
		s.mu.RLock()
		res = lo.Values(s.tenders)
		s.mu.RUnlock()
	}
	return trimSlice(res, limit, offset)
}

func (s *TenderStore) GetByID(id string) (*model.Tender, error) {
	var res *model.Tender
	lastVersion := defaultVersion
	s.mu.RLock()
	for _, tender := range s.tenders {
		if tender.ID == id {
			if tender.Version > lastVersion {
				lastVersion = tender.Version
				res = &tender
			}
		}
	}
	s.mu.RUnlock()
	if res == nil {
		return nil, TenderNotFound
	}
	return res, nil
}

func (s *TenderStore) GetByCreatorID(limit, offset int, creatorID string) []model.Tender {
	res := map[string]model.Tender{}
	s.mu.RLock()
	for _, t := range s.tenders {
		if t.CreatorID == creatorID {
			tend, exists := res[t.ID]
			if !exists {
				res[t.ID] = t
			} else {
				if t.Version > tend.Version {
					res[t.ID] = t
				}
			}
		}
	}
	values := lo.Values(res)
	s.mu.RUnlock()
	return trimSlice(values, limit, offset)
}

func (s *TenderStore) Save(t *model.Tender) error {
	tid := makeTid(t)
	s.mu.RLock()
	_, exists := s.tenders[tid]
	s.mu.RUnlock()
	if exists {
		slog.Debug("Tender already exists", "tender", t)
		return TenderAlreadyExists
	}
	s.mu.RUnlock()
	t.CreatedAt = time.Now()
	t.ID = uuid.New().String()
	t.Version = defaultVersion
	t.UpdatedAt = time.Now()
	s.mu.Lock()
	s.tenders[tid] = *t
	s.mu.Unlock()
	slog.Debug("Tender created", "tender", t)
	return nil
}

func (s *TenderStore) Update(t *model.Tender) error {
	tid := makeTid(t)
	s.mu.RLock()
	_, exists := s.tenders[tid]
	s.mu.RUnlock()
	if !exists {
		slog.Debug("Tender not found", "tender", t)
		return TenderNotFound
	}
	t.Version++
	newTid := makeTid(t)
	t.UpdatedAt = time.Now()
	s.mu.Lock()
	s.tenders[newTid] = *t
	s.mu.Unlock()
	slog.Debug("Tender updated", "tender", t)
	return nil
}

func (s *TenderStore) Rollback(id string, version int) (*model.Tender, error) {
	oldTid := tenderId{
		Id:  id,
		Ver: version,
	}

	s.mu.RLock()
	t, exists := s.tenders[oldTid]
	s.mu.RUnlock()
	if !exists {
		slog.Debug("Tender not found", "tender", t)
		return nil, TenderNotFound
	}
	newestT, err := s.GetByID(id)
	if err != nil {
		if errors.Is(TenderNotFound, err) {
			slog.Debug("Tender not found", "tender", t)
			return nil, TenderNotFound
		}
		slog.Error("Error getting tender", "error", err)
		return nil, err
	}
	t.Version = newestT.Version + 1
	t.UpdatedAt = time.Now()
	newTid := makeTid(&t)
	s.mu.Lock()
	s.tenders[newTid] = t
	s.mu.Unlock()
	slog.Debug("Tender rolled back", "tender", t)
	return &t, nil
}

func makeTid(tender *model.Tender) tenderId {
	return tenderId{
		Id:  tender.ID,
		Ver: tender.Version,
	}
}
