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

// TestDentistRepository_FindByID_ReturnsDefensiveCopy pins the
// copy-on-read/write invariant that prevents a data race between
// concurrent HTTP requests (e.g. a PUT and a GET for the same dentist
// ID) that would otherwise share and mutate the same *dentist.Dentist
// instance outside the repository's lock.
func TestDentistRepository_FindByID_ReturnsDefensiveCopy(t *testing.T) {
	repo := memory.NewDentistRepository()
	ctx := context.Background()
	d := newTestDentist(t, "id-1", "clinic-1")
	require.NoError(t, repo.Save(ctx, d))

	found, err := repo.FindByID(ctx, "id-1")
	require.NoError(t, err)

	// Mutating the caller's original pointer after Save must not affect
	// what the repository returns on a subsequent read.
	require.NoError(t, d.Update("Changed", "1", "changed@example.com", false, time.Now()))
	assert.NotEqual(t, "Changed", found.Name, "FindByID's returned copy should not be aliased by later mutations of the original pointer")

	// Mutating a previously returned copy must not affect a later read.
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
