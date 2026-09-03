package memory

import (
	"context"
	"fmt"
	"sync"

	"github.com/giancarlogoulart/capim-challenge-clinicas/internal/domain/dentist"
	"github.com/giancarlogoulart/capim-challenge-clinicas/internal/platform/apperrors"
)

// DentistRepository is a thread-safe in-memory implementation of
// dentist.Repository.
type DentistRepository struct {
	mu   sync.RWMutex
	data map[string]*dentist.Dentist
}

func NewDentistRepository() *DentistRepository {
	return &DentistRepository{data: make(map[string]*dentist.Dentist)}
}

func (r *DentistRepository) Save(_ context.Context, d *dentist.Dentist) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[d.ID] = d
	return nil
}

func (r *DentistRepository) FindByID(_ context.Context, id string) (*dentist.Dentist, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	d, ok := r.data[id]
	if !ok {
		return nil, apperrors.NotFound(fmt.Sprintf("dentist %s not found", id))
	}
	return d, nil
}

func (r *DentistRepository) FindByClinicID(_ context.Context, clinicID string) ([]*dentist.Dentist, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]*dentist.Dentist, 0)
	for _, d := range r.data {
		if d.ClinicID == clinicID {
			result = append(result, d)
		}
	}
	return result, nil
}

func (r *DentistRepository) Delete(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.data[id]; !ok {
		return apperrors.NotFound(fmt.Sprintf("dentist %s not found", id))
	}
	delete(r.data, id)
	return nil
}
