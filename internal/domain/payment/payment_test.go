package payment_test

import (
	"testing"
	"time"

	"github.com/giancarlogoulart/capim-challenge-clinicas/internal/domain/payment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPayment_Valid(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	amount, err := payment.NewMoney(500)
	require.NoError(t, err)

	p, err := payment.NewPayment("pay-1", "clinic-1", nil, amount, now)

	require.NoError(t, err)
	assert.Equal(t, payment.StatusPending, p.Status)
	assert.Equal(t, "", p.PixCode)
	assert.Nil(t, p.DentistID)
}

func TestNewPayment_WithOptionalDentist(t *testing.T) {
	dentistID := "dentist-1"
	amount, err := payment.NewMoney(500)
	require.NoError(t, err)

	p, err := payment.NewPayment("pay-1", "clinic-1", &dentistID, amount, time.Now())

	require.NoError(t, err)
	require.NotNil(t, p.DentistID)
	assert.Equal(t, dentistID, *p.DentistID)
}

func TestNewPayment_MissingClinicID(t *testing.T) {
	amount, err := payment.NewMoney(500)
	require.NoError(t, err)

	_, err = payment.NewPayment("pay-1", "", nil, amount, time.Now())
	assert.Error(t, err)
}

func TestPayment_SetPixCode(t *testing.T) {
	amount, err := payment.NewMoney(500)
	require.NoError(t, err)
	p, err := payment.NewPayment("pay-1", "clinic-1", nil, amount, time.Now())
	require.NoError(t, err)

	p.SetPixCode("00020126...")

	assert.Equal(t, "00020126...", p.PixCode)
}

func TestPayment_Approve_FromPending(t *testing.T) {
	created := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	approvedAt := created.Add(3 * time.Second)
	amount, err := payment.NewMoney(500)
	require.NoError(t, err)
	p, err := payment.NewPayment("pay-1", "clinic-1", nil, amount, created)
	require.NoError(t, err)

	err = p.Approve(approvedAt)

	require.NoError(t, err)
	assert.Equal(t, payment.StatusApproved, p.Status)
	assert.Equal(t, approvedAt, p.UpdatedAt)
}

func TestPayment_Approve_RejectsWhenAlreadyApproved(t *testing.T) {
	amount, err := payment.NewMoney(500)
	require.NoError(t, err)
	p, err := payment.NewPayment("pay-1", "clinic-1", nil, amount, time.Now())
	require.NoError(t, err)
	require.NoError(t, p.Approve(time.Now()))

	err = p.Approve(time.Now())

	assert.ErrorIs(t, err, payment.ErrInvalidTransition)
}
