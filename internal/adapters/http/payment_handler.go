package http

import (
	"context"
	"encoding/json"
	"net/http"

	paymentapp "github.com/giancarlogoulart/capim-challenge-clinicas/internal/application/payment"
	paymentdomain "github.com/giancarlogoulart/capim-challenge-clinicas/internal/domain/payment"
	"github.com/giancarlogoulart/capim-challenge-clinicas/internal/platform/apperrors"
)

// paymentService is the subset of paymentapp.Service the handler depends on —
// declared here (consumer side) so tests can supply a fake without needing
// the real application package.
type paymentService interface {
	Create(ctx context.Context, input paymentapp.CreateInput) (*paymentdomain.Payment, error)
	Get(ctx context.Context, id string) (*paymentdomain.Payment, error)
}

// PaymentHandler exposes HTTP handlers for the payment resource.
type PaymentHandler struct {
	service paymentService
}

// NewPaymentHandler builds a PaymentHandler backed by the given paymentService.
func NewPaymentHandler(service paymentService) *PaymentHandler {
	return &PaymentHandler{service: service}
}

func (h *PaymentHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req paymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, apperrors.Validation("invalid request body", map[string]string{"body": err.Error()}))
		return
	}
	p, err := h.service.Create(r.Context(), paymentapp.CreateInput{
		ClinicID:  req.ClinicID,
		DentistID: req.DentistID,
		Cents:     req.Cents,
	})
	if err != nil {
		WriteError(w, err)
		return
	}
	WriteJSON(w, http.StatusCreated, toPaymentResponse(p))
}

func (h *PaymentHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	p, err := h.service.Get(r.Context(), id)
	if err != nil {
		WriteError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, toPaymentResponse(p))
}
