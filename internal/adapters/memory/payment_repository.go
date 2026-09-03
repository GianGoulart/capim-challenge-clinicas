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

func (r *PaymentRepository) Save(_ context.Context, p *payment.Payment) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[p.ID] = p
	return nil
}

func (r *PaymentRepository) FindByID(_ context.Context, id string) (*payment.Payment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.data[id]
	if !ok {
		return nil, apperrors.NotFound(fmt.Sprintf("payment %s not found", id))
	}
	return p, nil
}
