package memory

import (
	"context"
	"fmt"
	"sync"

	"github.com/giancarlogoulart/capim-challenge-clinicas/internal/domain/payment"
	"github.com/giancarlogoulart/capim-challenge-clinicas/internal/platform/apperrors"
)

// PaymentRepository é uma implementação in-memory thread-safe de
// payment.Repository.
type PaymentRepository struct {
	mu   sync.RWMutex
	data map[string]*payment.Payment
}

// Asserção em tempo de compilação de que *PaymentRepository satisfaz
// payment.Repository. Veja a seção "Compile-time interface assertions"
// no README.md para a justificativa.
var _ payment.Repository = (*PaymentRepository)(nil)

func NewPaymentRepository() *PaymentRepository {
	return &PaymentRepository{data: make(map[string]*payment.Payment)}
}

// Save armazena uma cópia defensiva de p. Pagamentos são únicos entre os
// agregados deste pacote no sentido de que podem ser mutados
// concorrentemente por uma goroutine em background — o callback de
// confirmação assíncrona do PixProvider (veja
// application/payment.Service.onApproved) — enquanto um handler HTTP pode
// estar lendo o mesmo pagamento ao mesmo tempo. Armazenar (e retornar,
// veja FindByID) cópias em vez do ponteiro do chamador garante que
// nenhuma goroutine jamais compartilhe a mesma instância de
// *payment.Payment, o que de outra forma seria uma data race sobre seus
// campos Status/UpdatedAt.
func (r *PaymentRepository) Save(_ context.Context, p *payment.Payment) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	stored := *p
	r.data[p.ID] = &stored
	return nil
}

// FindByID retorna uma cópia defensiva do pagamento armazenado — veja o
// doc comment de Save para entender por que isso importa especialmente
// para este agregado.
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
