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

// Compile-time assertion that *ClinicRepository satisfies clinic.Repository.
// See the "Compile-time interface assertions" section in README.md for why
// this is worth having even though main.go's wiring would already catch a
// broken implementation at compile time.
var _ clinic.Repository = (*ClinicRepository)(nil)

func NewClinicRepository() *ClinicRepository {
	return &ClinicRepository{data: make(map[string]*clinic.Clinic)}
}

// Save stores a defensive copy of c. The application layer (see
// clinic.Service.Update/UpdateBankAccount) fetches a clinic via
// FindByID, mutates it in place, then calls Save again. Storing (and
// returning, see FindByID/FindAll) copies rather than the caller's
// pointer ensures no two goroutines — e.g. two concurrent HTTP
// requests, one PUT and one GET, for the same clinic ID — ever share
// the same *clinic.Clinic instance, which would otherwise be a data
// race on its fields.
func (r *ClinicRepository) Save(_ context.Context, c *clinic.Clinic) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	stored := *c
	r.data[c.ID] = &stored
	return nil
}

// FindByID returns a defensive copy of the stored clinic — see the Save
// doc comment for why this matters.
func (r *ClinicRepository) FindByID(_ context.Context, id string) (*clinic.Clinic, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	c, ok := r.data[id]
	if !ok {
		return nil, apperrors.NotFound(fmt.Sprintf("clinic %s not found", id))
	}
	found := *c
	return &found, nil
}

// FindAll returns defensive copies of all stored clinics — see the Save
// doc comment for why this matters.
func (r *ClinicRepository) FindAll(_ context.Context) ([]*clinic.Clinic, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]*clinic.Clinic, 0, len(r.data))
	for _, c := range r.data {
		found := *c
		result = append(result, &found)
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
