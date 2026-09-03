package memory

import (
	"context"
	"fmt"
	"sync"

	"github.com/giancarlogoulart/capim-challenge-clinicas/internal/domain/payment"
	"github.com/giancarlogoulart/capim-challenge-clinicas/internal/platform/apperrors"
)

// PaymentRepository is a thread-safe in-memory implementation of
// payment.Repository.
type PaymentRepository struct {
	mu   sync.RWMutex
	data map[string]*payment.Payment
}

func NewPaymentRepository() *PaymentRepository {
	return &PaymentRepository{data: make(map[string]*payment.Payment)}
}

// Save stores a defensive copy of p. Payments are unique among this
// package's aggregates in that they can be mutated concurrently from a
// background goroutine — the PixProvider's asynchronous confirmation
// callback (see application/payment.Service.onApproved) — while an HTTP
// handler may be reading the same payment at the same time. Storing (and
// returning, see FindByID) copies rather than the caller's pointer
// ensures no two goroutines ever share the same *payment.Payment
// instance, which would otherwise be a data race on its Status/UpdatedAt
// fields.
func (r *PaymentRepository) Save(_ context.Context, p *payment.Payment) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	stored := *p
	r.data[p.ID] = &stored
	return nil
}

// FindByID returns a defensive copy of the stored payment — see the Save
// doc comment for why this matters for this particular aggregate.
func (r *PaymentRepository) FindByID(_ context.Context, id string) (*payment.Payment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.data[id]
	if !ok {
		return nil, apperrors.NotFound(fmt.Sprintf("payment %s not found", id))
	}
	found := *p
	return &found, nil
}
