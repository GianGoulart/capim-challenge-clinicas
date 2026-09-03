package memory_test

import (
	"context"
	"testing"
	"time"

	"github.com/giancarlogoulart/capim-challenge-clinicas/internal/adapters/memory"
	"github.com/giancarlogoulart/capim-challenge-clinicas/internal/domain/payment"
	"github.com/giancarlogoulart/capim-challenge-clinicas/internal/platform/apperrors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestPayment(t *testing.T, id string) *payment.Payment {
	t.Helper()
	amount, err := payment.NewMoney(500)
	require.NoError(t, err)
	p, err := payment.NewPayment(id, "clinic-1", nil, amount, time.Now())
	require.NoError(t, err)
	return p
}

func TestPaymentRepository_SaveAndFindByID(t *testing.T) {
	repo := memory.NewPaymentRepository()
	ctx := context.Background()
	p := newTestPayment(t, "pay-1")

	require.NoError(t, repo.Save(ctx, p))

	found, err := repo.FindByID(ctx, "pay-1")
	require.NoError(t, err)
	assert.Equal(t, p, found)
}

func TestPaymentRepository_FindByID_NotFound(t *testing.T) {
	repo := memory.NewPaymentRepository()

	_, err := repo.FindByID(context.Background(), "missing")

	assert.True(t, apperrors.Is(err, apperrors.KindNotFound))
}

func TestPaymentRepository_SaveOverwritesExisting(t *testing.T) {
	repo := memory.NewPaymentRepository()
	ctx := context.Background()
	p := newTestPayment(t, "pay-1")
	require.NoError(t, repo.Save(ctx, p))

	require.NoError(t, p.Approve(time.Now()))
	require.NoError(t, repo.Save(ctx, p))

	found, err := repo.FindByID(ctx, "pay-1")
	require.NoError(t, err)
	assert.Equal(t, payment.StatusApproved, found.Status)
}

// TestPaymentRepository_FindByID_ReturnsDefensiveCopy pins the
// copy-on-read/write invariant that fixes the data race between the
// async Pix-approval goroutine and concurrent HTTP reads (see the doc
// comment on PaymentRepository.Save): the repository must never hand
// out a pointer that aliases its internal storage, in either
// direction.
func TestPaymentRepository_FindByID_ReturnsDefensiveCopy(t *testing.T) {
	repo := memory.NewPaymentRepository()
	ctx := context.Background()
	p := newTestPayment(t, "pay-1")
	require.NoError(t, repo.Save(ctx, p))

	found, err := repo.FindByID(ctx, "pay-1")
	require.NoError(t, err)

	// Mutating the caller's original pointer after Save must not affect
	// what the repository returns on a subsequent read.
	require.NoError(t, p.Approve(time.Now()))
	assert.NotEqual(t, payment.StatusApproved, found.Status, "FindByID's returned copy should not be aliased by later mutations of the original pointer")

	// Mutating a previously returned copy must not affect a later read.
	foundAgain, err := repo.FindByID(ctx, "pay-1")
	require.NoError(t, err)
	assert.NotSame(t, found, foundAgain, "each FindByID call must return a distinct copy")
	foundAgain.Status = payment.StatusApproved
	thirdRead, err := repo.FindByID(ctx, "pay-1")
	require.NoError(t, err)
	assert.NotEqual(t, payment.StatusApproved, thirdRead.Status, "mutating a returned copy must not leak into the repository's storage")
}
