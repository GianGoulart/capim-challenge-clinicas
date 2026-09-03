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
