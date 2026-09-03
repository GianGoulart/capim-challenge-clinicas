package memory

import (
	"context"
	"fmt"
	"sync"

	"github.com/giancarlogoulart/capim-challenge-clinicas/internal/domain/clinic"
	"github.com/giancarlogoulart/capim-challenge-clinicas/internal/platform/apperrors"
)

// ClinicRepository is a thread-safe in-memory implementation of
// clinic.Repository — no external database is used, per the challenge's
// technical requirements.
type ClinicRepository struct {
	mu   sync.RWMutex
	data map[string]*clinic.Clinic
}

func NewClinicRepository() *ClinicRepository {
	return &ClinicRepository{data: make(map[string]*clinic.Clinic)}
}

func (r *ClinicRepository) Save(_ context.Context, c *clinic.Clinic) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[c.ID] = c
	return nil
}

func (r *ClinicRepository) FindByID(_ context.Context, id string) (*clinic.Clinic, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	c, ok := r.data[id]
	if !ok {
		return nil, apperrors.NotFound(fmt.Sprintf("clinic %s not found", id))
	}
	return c, nil
}

func (r *ClinicRepository) FindAll(_ context.Context) ([]*clinic.Clinic, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]*clinic.Clinic, 0, len(r.data))
	for _, c := range r.data {
		result = append(result, c)
	}
	return result, nil
}

func (r *ClinicRepository) Delete(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.data[id]; !ok {
		return apperrors.NotFound(fmt.Sprintf("clinic %s not found", id))
	}
	delete(r.data, id)
	return nil
}
