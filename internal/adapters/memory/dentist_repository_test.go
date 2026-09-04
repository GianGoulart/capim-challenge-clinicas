package memory_test

import (
	"context"
	"testing"
	"time"

	"github.com/giancarlogoulart/capim-challenge-clinicas/internal/adapters/memory"
	"github.com/giancarlogoulart/capim-challenge-clinicas/internal/domain/dentist"
	"github.com/giancarlogoulart/capim-challenge-clinicas/internal/platform/apperrors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestDentist(t *testing.T, id, clinicID string) *dentist.Dentist {
	t.Helper()
	d, err := dentist.NewDentist(id, clinicID, "Dra. Ana", "+55 11 90000-0000", "ana@example.com", false, time.Now())
	require.NoError(t, err)
	return d
}

func TestDentistRepository_SaveAndFindByID(t *testing.T) {
	repo := memory.NewDentistRepository()
	ctx := context.Background()
	d := newTestDentist(t, "id-1", "clinic-1")

	require.NoError(t, repo.Save(ctx, d))

	found, err := repo.FindByID(ctx, "id-1")
	require.NoError(t, err)
	assert.Equal(t, d, found)
}

// TestDentistRepository_FindByID_ReturnsDefensiveCopy fixa o invariante de
// copy-on-read/write que previne uma data race entre requisições HTTP
// concorrentes (ex: um PUT e um GET para o mesmo ID de dentista) que, de
// outra forma, compartilhariam e mutariam a mesma instância de
// *dentist.Dentist fora do lock do repositório.
func TestDentistRepository_FindByID_ReturnsDefensiveCopy(t *testing.T) {
	repo := memory.NewDentistRepository()
	ctx := context.Background()
	d := newTestDentist(t, "id-1", "clinic-1")
	require.NoError(t, repo.Save(ctx, d))

	found, err := repo.FindByID(ctx, "id-1")
	require.NoError(t, err)

	// Mutar o ponteiro original do chamador após o Save não deve afetar
	// o que o repositório retorna numa leitura subsequente.
	require.NoError(t, d.Update("Changed", "1", "changed@example.com", false, time.Now()))
	assert.NotEqual(t, "Changed", found.Name, "FindByID's returned copy should not be aliased by later mutations of the original pointer")

	// Mutar uma cópia previamente retornada não deve afetar uma leitura posterior.
	foundAgain, err := repo.FindByID(ctx, "id-1")
	require.NoError(t, err)
	assert.NotSame(t, found, foundAgain, "each FindByID call must return a distinct copy")
	foundAgain.Name = "Mutated"
	thirdRead, err := repo.FindByID(ctx, "id-1")
	require.NoError(t, err)
	assert.NotEqual(t, "Mutated", thirdRead.Name, "mutating a returned copy must not leak into the repository's storage")
}

func TestDentistRepository_FindByID_NotFound(t *testing.T) {
	repo := memory.NewDentistRepository()

	_, err := repo.FindByID(context.Background(), "missing")

	assert.True(t, apperrors.Is(err, apperrors.KindNotFound))
}

func TestDentistRepository_FindByClinicID(t *testing.T) {
	repo := memory.NewDentistRepository()
	ctx := context.Background()
	require.NoError(t, repo.Save(ctx, newTestDentist(t, "id-1", "clinic-1")))
	require.NoError(t, repo.Save(ctx, newTestDentist(t, "id-2", "clinic-1")))
	require.NoError(t, repo.Save(ctx, newTestDentist(t, "id-3", "clinic-2")))

	found, err := repo.FindByClinicID(ctx, "clinic-1")

	require.NoError(t, err)
	assert.Len(t, found, 2)
}

func TestDentistRepository_Delete(t *testing.T) {
	repo := memory.NewDentistRepository()
	ctx := context.Background()
	require.NoError(t, repo.Save(ctx, newTestDentist(t, "id-1", "clinic-1")))

	require.NoError(t, repo.Delete(ctx, "id-1"))

	_, err := repo.FindByID(ctx, "id-1")
	assert.True(t, apperrors.Is(err, apperrors.KindNotFound))
}
