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

// Save stores a defensive copy of d. The application layer (see
// dentist.Service.Update) fetches a dentist via FindByID, mutates it in
// place, then calls Save again. Storing (and returning, see
// FindByID/FindByClinicID) copies rather than the caller's pointer
// ensures no two goroutines — e.g. two concurrent HTTP requests, one
// PUT and one GET, for the same dentist ID — ever share the same
// *dentist.Dentist instance, which would otherwise be a data race on
// its fields.
func (r *DentistRepository) Save(_ context.Context, d *dentist.Dentist) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	stored := *d
	r.data[d.ID] = &stored
	return nil
}

// FindByID returns a defensive copy of the stored dentist — see the
// Save doc comment for why this matters.
func (r *DentistRepository) FindByID(_ context.Context, id string) (*dentist.Dentist, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	d, ok := r.data[id]
	if !ok {
		return nil, apperrors.NotFound(fmt.Sprintf("dentist %s not found", id))
	}
	found := *d
	return &found, nil
}

// FindByClinicID returns defensive copies of the matching dentists —
// see the Save doc comment for why this matters.
func (r *DentistRepository) FindByClinicID(_ context.Context, clinicID string) ([]*dentist.Dentist, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]*dentist.Dentist, 0)
	for _, d := range r.data {
		if d.ClinicID == clinicID {
			found := *d
			result = append(result, &found)
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
