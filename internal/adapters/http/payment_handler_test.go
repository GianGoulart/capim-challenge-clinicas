package http_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	httpadapter "github.com/giancarlogoulart/capim-challenge-clinicas/internal/adapters/http"
	paymentapp "github.com/giancarlogoulart/capim-challenge-clinicas/internal/application/payment"
	paymentdomain "github.com/giancarlogoulart/capim-challenge-clinicas/internal/domain/payment"
	"github.com/giancarlogoulart/capim-challenge-clinicas/internal/platform/apperrors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakePaymentService struct {
	createFn func(ctx context.Context, input paymentapp.CreateInput) (*paymentdomain.Payment, error)
	getFn    func(ctx context.Context, id string) (*paymentdomain.Payment, error)
}

func (f *fakePaymentService) Create(ctx context.Context, input paymentapp.CreateInput) (*paymentdomain.Payment, error) {
	return f.createFn(ctx, input)
}
func (f *fakePaymentService) Get(ctx context.Context, id string) (*paymentdomain.Payment, error) {
	return f.getFn(ctx, id)
}

func samplePayment(t *testing.T) *paymentdomain.Payment {
	t.Helper()
	amount, err := paymentdomain.NewMoney(1000)
	require.NoError(t, err)
	p, err := paymentdomain.NewPayment("pay-1", "clinic-1", nil, amount, time.Now())
	require.NoError(t, err)
	p.SetPixCode("FAKE-PIX-CODE")
	return p
}

func TestPaymentHandler_Create_Success(t *testing.T) {
	svc := &fakePaymentService{
		createFn: func(ctx context.Context, input paymentapp.CreateInput) (*paymentdomain.Payment, error) {
			return samplePayment(t), nil
		},
	}
	handler := httpadapter.NewPaymentHandler(svc)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/payments", bytes.NewBufferString(`{"clinic_id":"clinic-1","amount_cents":1000}`))
	rec := httptest.NewRecorder()

	handler.Create(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
}

func TestPaymentHandler_Create_ClinicNotFound(t *testing.T) {
	svc := &fakePaymentService{
		createFn: func(ctx context.Context, input paymentapp.CreateInput) (*paymentdomain.Payment, error) {
			return nil, apperrors.NotFound("clinic not found")
		},
	}
	handler := httpadapter.NewPaymentHandler(svc)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/payments", bytes.NewBufferString(`{"clinic_id":"missing","amount_cents":1000}`))
	rec := httptest.NewRecorder()

	handler.Create(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestPaymentHandler_Create_InvalidAmount(t *testing.T) {
	svc := &fakePaymentService{
		createFn: func(ctx context.Context, input paymentapp.CreateInput) (*paymentdomain.Payment, error) {
			return nil, apperrors.Validation("invalid amount", nil)
		},
	}
	handler := httpadapter.NewPaymentHandler(svc)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/payments", bytes.NewBufferString(`{"clinic_id":"clinic-1","amount_cents":0}`))
	rec := httptest.NewRecorder()

	handler.Create(rec, req)

	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestPaymentHandler_Get_Success(t *testing.T) {
	svc := &fakePaymentService{
		getFn: func(ctx context.Context, id string) (*paymentdomain.Payment, error) { return samplePayment(t), nil },
	}
	handler := httpadapter.NewPaymentHandler(svc)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/payments/pay-1", nil)
	req.SetPathValue("id", "pay-1")
	rec := httptest.NewRecorder()

	handler.Get(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestPaymentHandler_Get_NotFound(t *testing.T) {
	svc := &fakePaymentService{
		getFn: func(ctx context.Context, id string) (*paymentdomain.Payment, error) {
			return nil, apperrors.NotFound("payment not found")
		},
	}
	handler := httpadapter.NewPaymentHandler(svc)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/payments/missing", nil)
	req.SetPathValue("id", "missing")
	rec := httptest.NewRecorder()

	handler.Get(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}
