package http

import (
	"context"
	"encoding/json"
	"net/http"

	clinicapp "github.com/giancarlogoulart/capim-challenge-clinicas/internal/application/clinic"
	clinicdomain "github.com/giancarlogoulart/capim-challenge-clinicas/internal/domain/clinic"
	"github.com/giancarlogoulart/capim-challenge-clinicas/internal/platform/apperrors"
)

// clinicService is the subset of clinicapp.Service the handler depends on —
// declared here (consumer side) so tests can supply a fake without needing
// the real application package.
type clinicService interface {
	Create(ctx context.Context, input clinicapp.CreateInput) (*clinicdomain.Clinic, error)
	Get(ctx context.Context, id string) (*clinicdomain.Clinic, error)
	List(ctx context.Context) ([]*clinicdomain.Clinic, error)
	Update(ctx context.Context, id string, input clinicapp.UpdateInput) (*clinicdomain.Clinic, error)
	UpdateBankAccount(ctx context.Context, id string, input clinicapp.UpdateBankAccountInput) (*clinicdomain.Clinic, error)
	Delete(ctx context.Context, id string) error
}

// ClinicHandler exposes HTTP handlers for the clinic resource.
type ClinicHandler struct {
	service clinicService
}

// NewClinicHandler builds a ClinicHandler backed by the given clinicService.
func NewClinicHandler(service clinicService) *ClinicHandler {
	return &ClinicHandler{service: service}
}

func (h *ClinicHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req clinicRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, apperrors.Validation("invalid request body", map[string]string{"body": err.Error()}))
		return
	}
	c, err := h.service.Create(r.Context(), clinicapp.CreateInput{
		Document:      req.Document,
		CorporateName: req.CorporateName,
		TradeName:     req.TradeName,
		BankCode:      req.BankCode,
		Agency:        req.Agency,
		Account:       req.Account,
	})
	if err != nil {
		WriteError(w, err)
		return
	}
	WriteJSON(w, http.StatusCreated, toClinicResponse(c))
}

func (h *ClinicHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	c, err := h.service.Get(r.Context(), id)
	if err != nil {
		WriteError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, toClinicResponse(c))
}

func (h *ClinicHandler) List(w http.ResponseWriter, r *http.Request) {
	clinics, err := h.service.List(r.Context())
	if err != nil {
		WriteError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, toClinicResponseList(clinics))
}

func (h *ClinicHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req clinicUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, apperrors.Validation("invalid request body", map[string]string{"body": err.Error()}))
		return
	}
	c, err := h.service.Update(r.Context(), id, clinicapp.UpdateInput{
		CorporateName: req.CorporateName,
		TradeName:     req.TradeName,
	})
	if err != nil {
		WriteError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, toClinicResponse(c))
}

func (h *ClinicHandler) UpdateBankAccount(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req bankAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, apperrors.Validation("invalid request body", map[string]string{"body": err.Error()}))
		return
	}
	c, err := h.service.UpdateBankAccount(r.Context(), id, clinicapp.UpdateBankAccountInput{
		BankCode: req.BankCode,
		Agency:   req.Agency,
		Account:  req.Account,
	})
	if err != nil {
		WriteError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, toClinicResponse(c))
}

func (h *ClinicHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.service.Delete(r.Context(), id); err != nil {
		WriteError(w, err)
		return
	}
	WriteJSON(w, http.StatusNoContent, nil)
}
