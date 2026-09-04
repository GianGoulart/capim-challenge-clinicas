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

// TestPaymentRepository_FindByID_ReturnsDefensiveCopy fixa o invariante de
// copy-on-read/write que corrige a data race entre a goroutine assíncrona
// de aprovação do Pix e leituras HTTP concorrentes (veja o doc comment de
// PaymentRepository.Save): o repositório nunca deve entregar um ponteiro
// que compartilhe alias com seu armazenamento interno, em nenhuma das
// direções.
func TestPaymentRepository_FindByID_ReturnsDefensiveCopy(t *testing.T) {
	repo := memory.NewPaymentRepository()
	ctx := context.Background()
	p := newTestPayment(t, "pay-1")
	require.NoError(t, repo.Save(ctx, p))

	found, err := repo.FindByID(ctx, "pay-1")
	require.NoError(t, err)

	// Mutar o ponteiro original do chamador após o Save não deve afetar
	// o que o repositório retorna numa leitura subsequente.
	require.NoError(t, p.Approve(time.Now()))
	assert.NotEqual(t, payment.StatusApproved, found.Status, "FindByID's returned copy should not be aliased by later mutations of the original pointer")

	// Mutar uma cópia previamente retornada não deve afetar uma leitura posterior.
	foundAgain, err := repo.FindByID(ctx, "pay-1")
	require.NoError(t, err)
	assert.NotSame(t, found, foundAgain, "each FindByID call must return a distinct copy")
	foundAgain.Status = payment.StatusApproved
	thirdRead, err := repo.FindByID(ctx, "pay-1")
	require.NoError(t, err)
	assert.NotEqual(t, payment.StatusApproved, thirdRead.Status, "mutating a returned copy must not leak into the repository's storage")
}
