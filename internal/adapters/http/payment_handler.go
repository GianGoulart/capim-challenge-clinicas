package http

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/giancarlogoulart/capim-challenge-clinicas/internal/adapters/http/dto"
	paymentapp "github.com/giancarlogoulart/capim-challenge-clinicas/internal/application/payment"
	paymentdomain "github.com/giancarlogoulart/capim-challenge-clinicas/internal/domain/payment"
	"github.com/giancarlogoulart/capim-challenge-clinicas/internal/platform/apperrors"
)

// paymentService é o subconjunto de paymentapp.Service do qual o handler depende —
// declarado aqui (consumer side) para que os testes possam fornecer um fake sem precisar
// do pacote de application real.
type paymentService interface {
	Create(ctx context.Context, input paymentapp.CreateInput) (*paymentdomain.Payment, error)
	Get(ctx context.Context, id string) (*paymentdomain.Payment, error)
}

// PaymentHandler expõe os handlers HTTP para o recurso de pagamento.
type PaymentHandler struct {
	service paymentService
}

// NewPaymentHandler cria um PaymentHandler baseado no paymentService informado.
func NewPaymentHandler(service paymentService) *PaymentHandler {
	return &PaymentHandler{service: service}
}

func (h *PaymentHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.PaymentRequest
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
	WriteJSON(w, http.StatusCreated, dto.ToPaymentResponse(p))
}

func (h *PaymentHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	p, err := h.service.Get(r.Context(), id)
	if err != nil {
		WriteError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, dto.ToPaymentResponse(p))
}
