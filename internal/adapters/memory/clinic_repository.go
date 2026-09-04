package memory

import (
	"context"
	"fmt"
	"sync"

	"github.com/giancarlogoulart/capim-challenge-clinicas/internal/domain/clinic"
	"github.com/giancarlogoulart/capim-challenge-clinicas/internal/platform/apperrors"
)

// ClinicRepository é uma implementação in-memory thread-safe de
// clinic.Repository — nenhum banco de dados externo é usado, conforme os
// requisitos técnicos do desafio.
type ClinicRepository struct {
	mu   sync.RWMutex
	data map[string]*clinic.Clinic
}

// Asserção em tempo de compilação de que *ClinicRepository satisfaz
// clinic.Repository. Veja a seção "Compile-time interface assertions" no
// README.md para entender por que vale a pena ter isso mesmo que o wiring
// do main.go já pegasse uma implementação quebrada em tempo de compilação.
var _ clinic.Repository = (*ClinicRepository)(nil)

func NewClinicRepository() *ClinicRepository {
	return &ClinicRepository{data: make(map[string]*clinic.Clinic)}
}

// Save armazena uma cópia defensiva de c. A camada de application (veja
// clinic.Service.Update/UpdateBankAccount) busca uma clínica via
// FindByID, muta ela em memória, e então chama Save de novo. Armazenar (e
// retornar, veja FindByID/FindAll) cópias em vez do ponteiro do chamador
// garante que nenhuma goroutine — por exemplo, duas requisições HTTP
// concorrentes, um PUT e um GET, para o mesmo ID de clínica — jamais
// compartilhe a mesma instância de *clinic.Clinic, o que de outra forma
// seria uma data race sobre seus campos.
func (r *ClinicRepository) Save(_ context.Context, c *clinic.Clinic) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	stored := *c
	r.data[c.ID] = &stored
	return nil
}

// FindByID retorna uma cópia defensiva da clínica armazenada — veja o
// doc comment de Save para entender por que isso importa.
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

// FindAll retorna cópias defensivas de todas as clínicas armazenadas — veja
// o doc comment de Save para entender por que isso importa.
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
