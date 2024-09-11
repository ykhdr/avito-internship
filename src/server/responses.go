package server

import (
	"fmt"
	"time"
)

type JSONTime time.Time

func (t JSONTime) MarshalJSON() ([]byte, error) {
	stamp := fmt.Sprintf("\"%s\"", time.Time(t).Format(time.RFC3339))
	return []byte(stamp), nil
}

type ErrResponse struct {
	Reason string `json:"reason"`
}

type TenderResponse struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Status      string   `json:"status"`
	ServiceType string   `json:"serviceType"`
	Version     int      `json:"version"`
	CreatedAt   JSONTime `json:"createdAt"`
}
