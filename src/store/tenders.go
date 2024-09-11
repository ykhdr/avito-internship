package store

import (
	"github.com/google/uuid"
	"log/slog"
	"slices"
	"time"
	"zadanie-6105/model"
)
import "github.com/samber/lo"

type TenderStore struct {
	tenders []*model.Tender
}

func NewTendersStore() *TenderStore {
	return &TenderStore{}
}

func (s *TenderStore) GetAll(limit, offset int, serviceType []string) []*model.Tender {
	var res []*model.Tender
	if len(serviceType) != 0 {
		res = lo.Filter(s.tenders, func(t *model.Tender, _ int) bool {
			if slices.Contains(serviceType, t.ServiceType) {
				return true
			}
			return false
		})
	} else {
		res = s.tenders
	}
	return trimSlice(res, limit, offset)
}

func (s *TenderStore) GetByID(id string) *model.Tender {
	for _, tender := range s.tenders {
		if tender.ID == id {
			return tender
		}
	}
	return nil
}

func (s *TenderStore) GetByCreatorID(limit, offset int, creatorID string) []*model.Tender {
	res := lo.Filter(s.tenders, func(t *model.Tender, _ int) bool {
		return t.CreatorID == creatorID
	})
	return trimSlice(res, limit, offset)
}

func (s *TenderStore) Save(t *model.Tender) {
	t.CreatedAt = time.Now()
	t.UpdatedAt = time.Now()
	t.ID = uuid.New().String()
	s.tenders = append(s.tenders, t)
	slog.Debug("Tender created", "tender", t)
}

func (s *TenderStore) Update(t *model.Tender) {
	for i, tender := range s.tenders {
		if tender.ID == t.ID {
			t.UpdatedAt = time.Now()
			s.tenders[i] = t
			slog.Debug("Tender updated", "tender", t)
			return
		}
	}
	slog.Debug("Tender not found", "tender", t)
}
