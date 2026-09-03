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

// TestClinicRepository_FindByID_ReturnsDefensiveCopy pins the
// copy-on-read/write invariant that prevents a data race between
// concurrent HTTP requests (e.g. a PUT and a GET for the same clinic
// ID) that would otherwise share and mutate the same *clinic.Clinic
// instance outside the repository's lock.
func TestClinicRepository_FindByID_ReturnsDefensiveCopy(t *testing.T) {
	repo := memory.NewClinicRepository()
	ctx := context.Background()
	c := newTestClinic(t, "id-1")
	require.NoError(t, repo.Save(ctx, c))

	found, err := repo.FindByID(ctx, "id-1")
	require.NoError(t, err)

	// Mutating the caller's original pointer after Save must not affect
	// what the repository returns on a subsequent read.
	require.NoError(t, c.UpdateInfo("Changed", "Changed", time.Now()))
	assert.NotEqual(t, "Changed", found.CorporateName, "FindByID's returned copy should not be aliased by later mutations of the original pointer")

	// Mutating a previously returned copy must not affect a later read.
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
