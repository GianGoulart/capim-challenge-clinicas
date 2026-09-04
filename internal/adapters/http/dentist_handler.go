package http

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/giancarlogoulart/capim-challenge-clinicas/internal/adapters/http/dto"
	dentistapp "github.com/giancarlogoulart/capim-challenge-clinicas/internal/application/dentist"
	dentistdomain "github.com/giancarlogoulart/capim-challenge-clinicas/internal/domain/dentist"
	"github.com/giancarlogoulart/capim-challenge-clinicas/internal/platform/apperrors"
)

// dentistService is the subset of dentistapp.Service the handler depends on —
// declared here (consumer side) so tests can supply a fake without needing
// the real application package.
type dentistService interface {
	Create(ctx context.Context, input dentistapp.CreateInput) (*dentistdomain.Dentist, error)
	Get(ctx context.Context, id string) (*dentistdomain.Dentist, error)
	ListByClinic(ctx context.Context, clinicID string) ([]*dentistdomain.Dentist, error)
	Update(ctx context.Context, id string, input dentistapp.UpdateInput) (*dentistdomain.Dentist, error)
	Delete(ctx context.Context, id string) error
}

// DentistHandler exposes HTTP handlers for the dentist resource.
type DentistHandler struct {
	service dentistService
}

// NewDentistHandler builds a DentistHandler backed by the given dentistService.
func NewDentistHandler(service dentistService) *DentistHandler {
	return &DentistHandler{service: service}
}

func (h *DentistHandler) Create(w http.ResponseWriter, r *http.Request) {
	clinicID := r.PathValue("clinic_id")
	var req dto.DentistRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, apperrors.Validation("invalid request body", map[string]string{"body": err.Error()}))
		return
	}
	d, err := h.service.Create(r.Context(), dentistapp.CreateInput{
		ClinicID: clinicID,
		Name:     req.Name,
		Phone:    req.Phone,
		Email:    req.Email,
		IsAdmin:  req.IsAdmin,
	})
	if err != nil {
		WriteError(w, err)
		return
	}
	WriteJSON(w, http.StatusCreated, dto.ToDentistResponse(d))
}

func (h *DentistHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	d, err := h.service.Get(r.Context(), id)
	if err != nil {
		WriteError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, dto.ToDentistResponse(d))
}

func (h *DentistHandler) ListByClinic(w http.ResponseWriter, r *http.Request) {
	clinicID := r.PathValue("clinic_id")
	dentists, err := h.service.ListByClinic(r.Context(), clinicID)
	if err != nil {
		WriteError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, dto.ToDentistResponseList(dentists))
}

func (h *DentistHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req dto.DentistRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, apperrors.Validation("invalid request body", map[string]string{"body": err.Error()}))
		return
	}
	d, err := h.service.Update(r.Context(), id, dentistapp.UpdateInput{
		Name:    req.Name,
		Phone:   req.Phone,
		Email:   req.Email,
		IsAdmin: req.IsAdmin,
	})
	if err != nil {
		WriteError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, dto.ToDentistResponse(d))
}

func (h *DentistHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.service.Delete(r.Context(), id); err != nil {
		WriteError(w, err)
		return
	}
	WriteJSON(w, http.StatusNoContent, nil)
}
