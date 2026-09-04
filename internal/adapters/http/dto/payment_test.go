package dto_test

import (
	"testing"
	"time"

	"github.com/giancarlogoulart/capim-challenge-clinicas/internal/adapters/http/dto"
	paymentdomain "github.com/giancarlogoulart/capim-challenge-clinicas/internal/domain/payment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func samplePayment(t *testing.T, dentistID *string) *paymentdomain.Payment {
	t.Helper()
	amount, err := paymentdomain.NewMoney(1500)
	require.NoError(t, err)
	p, err := paymentdomain.NewPayment("payment-1", "clinic-1", dentistID, amount, time.Now())
	require.NoError(t, err)
	p.SetPixCode("FAKE-PIX-CODE")
	return p
}

func TestToPaymentResponse_MapsAllFields(t *testing.T) {
	p := samplePayment(t, nil)

	resp := dto.ToPaymentResponse(p)

	assert.Equal(t, "payment-1", resp.ID)
	assert.Equal(t, "clinic-1", resp.ClinicID)
	assert.Nil(t, resp.DentistID)
	assert.Equal(t, int64(1500), resp.AmountCents)
	assert.Equal(t, "pending", resp.Status)
	assert.Equal(t, "FAKE-PIX-CODE", resp.PixCode)
}

func TestToPaymentResponse_IncludesDentistIDWhenPresent(t *testing.T) {
	dentistID := "dentist-1"

	resp := dto.ToPaymentResponse(samplePayment(t, &dentistID))

	require.NotNil(t, resp.DentistID)
	assert.Equal(t, "dentist-1", *resp.DentistID)
}

func TestToPaymentResponse_ReflectsApprovedStatus(t *testing.T) {
	p := samplePayment(t, nil)
	require.NoError(t, p.Approve(time.Now()))

	resp := dto.ToPaymentResponse(p)

	assert.Equal(t, "approved", resp.Status)
}
