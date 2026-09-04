package memory_test

import (
	"context"
	"testing"
	"time"

	"github.com/giancarlogoulart/capim-challenge-clinicas/internal/adapters/memory"
	"github.com/giancarlogoulart/capim-challenge-clinicas/internal/domain/clinic"
	"github.com/giancarlogoulart/capim-challenge-clinicas/internal/platform/apperrors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestClinic(t *testing.T, id string) *clinic.Clinic {
	t.Helper()
	doc, err := clinic.NewDocument("52998224725")
	require.NoError(t, err)
	acc, err := clinic.NewBankAccount("341", "1234", "56789-0")
	require.NoError(t, err)
	c, err := clinic.NewClinic(id, doc, "Corp", "Trade", acc, time.Now())
	require.NoError(t, err)
	return c
}

func TestClinicRepository_SaveAndFindByID(t *testing.T) {
	repo := memory.NewClinicRepository()
	ctx := context.Background()
	c := newTestClinic(t, "id-1")

	require.NoError(t, repo.Save(ctx, c))

	found, err := repo.FindByID(ctx, "id-1")
	require.NoError(t, err)
	assert.Equal(t, c, found)
}

// TestClinicRepository_FindByID_ReturnsDefensiveCopy fixa o invariante de
// copy-on-read/write que previne uma data race entre requisições HTTP
// concorrentes (ex: um PUT e um GET para o mesmo ID de clínica) que, de
// outra forma, compartilhariam e mutariam a mesma instância de
// *clinic.Clinic fora do lock do repositório.
func TestClinicRepository_FindByID_ReturnsDefensiveCopy(t *testing.T) {
	repo := memory.NewClinicRepository()
	ctx := context.Background()
	c := newTestClinic(t, "id-1")
	require.NoError(t, repo.Save(ctx, c))

	found, err := repo.FindByID(ctx, "id-1")
	require.NoError(t, err)

	// Mutar o ponteiro original do chamador após o Save não deve afetar
	// o que o repositório retorna numa leitura subsequente.
	require.NoError(t, c.UpdateInfo("Changed", "Changed", time.Now()))
	assert.NotEqual(t, "Changed", found.CorporateName, "FindByID's returned copy should not be aliased by later mutations of the original pointer")

	// Mutar uma cópia previamente retornada não deve afetar uma leitura posterior.
	foundAgain, err := repo.FindByID(ctx, "id-1")
	require.NoError(t, err)
	assert.NotSame(t, found, foundAgain, "each FindByID call must return a distinct copy")
	foundAgain.CorporateName = "Mutated"
	thirdRead, err := repo.FindByID(ctx, "id-1")
	require.NoError(t, err)
	assert.NotEqual(t, "Mutated", thirdRead.CorporateName, "mutating a returned copy must not leak into the repository's storage")
}

func TestClinicRepository_FindByID_NotFound(t *testing.T) {
	repo := memory.NewClinicRepository()

	_, err := repo.FindByID(context.Background(), "missing")

	assert.True(t, apperrors.Is(err, apperrors.KindNotFound))
}

func TestClinicRepository_FindAll(t *testing.T) {
	repo := memory.NewClinicRepository()
	ctx := context.Background()
	require.NoError(t, repo.Save(ctx, newTestClinic(t, "id-1")))
	require.NoError(t, repo.Save(ctx, newTestClinic(t, "id-2")))

	all, err := repo.FindAll(ctx)

	require.NoError(t, err)
	assert.Len(t, all, 2)
}

func TestClinicRepository_Delete(t *testing.T) {
	repo := memory.NewClinicRepository()
	ctx := context.Background()
	require.NoError(t, repo.Save(ctx, newTestClinic(t, "id-1")))

	require.NoError(t, repo.Delete(ctx, "id-1"))

	_, err := repo.FindByID(ctx, "id-1")
	assert.True(t, apperrors.Is(err, apperrors.KindNotFound))
}

func TestClinicRepository_Delete_NotFound(t *testing.T) {
	repo := memory.NewClinicRepository()

	err := repo.Delete(context.Background(), "missing")

	assert.True(t, apperrors.Is(err, apperrors.KindNotFound))
}
