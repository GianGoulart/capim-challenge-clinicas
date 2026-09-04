package memory

import (
	"context"
	"fmt"
	"sync"

	"github.com/giancarlogoulart/capim-challenge-clinicas/internal/domain/dentist"
	"github.com/giancarlogoulart/capim-challenge-clinicas/internal/platform/apperrors"
)

// DentistRepository é uma implementação in-memory thread-safe de
// dentist.Repository.
type DentistRepository struct {
	mu   sync.RWMutex
	data map[string]*dentist.Dentist
}

// Asserção em tempo de compilação de que *DentistRepository satisfaz dentist.Repository
var _ dentist.Repository = (*DentistRepository)(nil)

func NewDentistRepository() *DentistRepository {
	return &DentistRepository{data: make(map[string]*dentist.Dentist)}
}

// Save armazena uma cópia defensiva de d. A camada de application (veja
// dentist.Service.Update) busca um dentista via FindByID, muta ele em
// memória, e então chama Save de novo. Armazenar (e retornar, veja
// FindByID/FindByClinicID) cópias em vez do ponteiro do chamador garante
// que nenhuma goroutine — por exemplo, duas requisições HTTP
// concorrentes, um PUT e um GET, para o mesmo ID de dentista — jamais
// compartilhe a mesma instância de *dentist.Dentist, o que de outra forma
// seria uma data race sobre seus campos.
func (r *DentistRepository) Save(_ context.Context, d *dentist.Dentist) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	stored := *d
	r.data[d.ID] = &stored
	return nil
}

// FindByID retorna uma cópia defensiva do dentista armazenado — veja o
// doc comment de Save para entender por que isso importa.
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

// FindByClinicID retorna cópias defensivas dos dentistas correspondentes —
// veja o doc comment de Save para entender por que isso importa.
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
