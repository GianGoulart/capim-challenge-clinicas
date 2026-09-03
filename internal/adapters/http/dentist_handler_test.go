package http_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	httpadapter "github.com/giancarlogoulart/capim-challenge-clinicas/internal/adapters/http"
	dentistapp "github.com/giancarlogoulart/capim-challenge-clinicas/internal/application/dentist"
	dentistdomain "github.com/giancarlogoulart/capim-challenge-clinicas/internal/domain/dentist"
	"github.com/giancarlogoulart/capim-challenge-clinicas/internal/platform/apperrors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeDentistService struct {
	createFn func(ctx context.Context, input dentistapp.CreateInput) (*dentistdomain.Dentist, error)
	getFn    func(ctx context.Context, id string) (*dentistdomain.Dentist, error)
	listFn   func(ctx context.Context, clinicID string) ([]*dentistdomain.Dentist, error)
	updateFn func(ctx context.Context, id string, input dentistapp.UpdateInput) (*dentistdomain.Dentist, error)
	deleteFn func(ctx context.Context, id string) error
}

func (f *fakeDentistService) Create(ctx context.Context, input dentistapp.CreateInput) (*dentistdomain.Dentist, error) {
	return f.createFn(ctx, input)
}
func (f *fakeDentistService) Get(ctx context.Context, id string) (*dentistdomain.Dentist, error) {
	return f.getFn(ctx, id)
}
func (f *fakeDentistService) ListByClinic(ctx context.Context, clinicID string) ([]*dentistdomain.Dentist, error) {
	return f.listFn(ctx, clinicID)
}
func (f *fakeDentistService) Update(ctx context.Context, id string, input dentistapp.UpdateInput) (*dentistdomain.Dentist, error) {
	return f.updateFn(ctx, id, input)
}
func (f *fakeDentistService) Delete(ctx context.Context, id string) error {
	return f.deleteFn(ctx, id)
}

func sampleDentist(t *testing.T) *dentistdomain.Dentist {
	t.Helper()
	d, err := dentistdomain.NewDentist("dentist-1", "clinic-1", "Dra. Ana", "+55 11 90000-0000", "ana@example.com", true, timeNow())
	require.NoError(t, err)
	return d
}

func TestDentistHandler_Create_Success(t *testing.T) {
	svc := &fakeDentistService{
		createFn: func(ctx context.Context, input dentistapp.CreateInput) (*dentistdomain.Dentist, error) {
			return sampleDentist(t), nil
		},
	}
	handler := httpadapter.NewDentistHandler(svc)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/clinics/clinic-1/dentists", bytes.NewBufferString(`{"name":"Dra. Ana","phone":"+55 11 90000-0000","email":"ana@example.com","is_admin":true}`))
	req.SetPathValue("clinic_id", "clinic-1")
	rec := httptest.NewRecorder()

	handler.Create(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
}

func TestDentistHandler_Create_ClinicNotFound(t *testing.T) {
	svc := &fakeDentistService{
		createFn: func(ctx context.Context, input dentistapp.CreateInput) (*dentistdomain.Dentist, error) {
			return nil, apperrors.NotFound("clinic not found")
		},
	}
	handler := httpadapter.NewDentistHandler(svc)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/clinics/missing/dentists", bytes.NewBufferString(`{}`))
	req.SetPathValue("clinic_id", "missing")
	rec := httptest.NewRecorder()

	handler.Create(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestDentistHandler_Get_Success(t *testing.T) {
	svc := &fakeDentistService{
		getFn: func(ctx context.Context, id string) (*dentistdomain.Dentist, error) { return sampleDentist(t), nil },
	}
	handler := httpadapter.NewDentistHandler(svc)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/dentists/dentist-1", nil)
	req.SetPathValue("id", "dentist-1")
	rec := httptest.NewRecorder()

	handler.Get(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestDentistHandler_ListByClinic_Success(t *testing.T) {
	svc := &fakeDentistService{
		listFn: func(ctx context.Context, clinicID string) ([]*dentistdomain.Dentist, error) {
			return []*dentistdomain.Dentist{sampleDentist(t)}, nil
		},
	}
	handler := httpadapter.NewDentistHandler(svc)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/clinics/clinic-1/dentists", nil)
	req.SetPathValue("clinic_id", "clinic-1")
	rec := httptest.NewRecorder()

	handler.ListByClinic(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestDentistHandler_Update_Success(t *testing.T) {
	svc := &fakeDentistService{
		updateFn: func(ctx context.Context, id string, input dentistapp.UpdateInput) (*dentistdomain.Dentist, error) {
			return sampleDentist(t), nil
		},
	}
	handler := httpadapter.NewDentistHandler(svc)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/dentists/dentist-1", bytes.NewBufferString(`{"name":"Dra. Ana","phone":"1","email":"a@a.com"}`))
	req.SetPathValue("id", "dentist-1")
	rec := httptest.NewRecorder()

	handler.Update(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestDentistHandler_Delete_Success(t *testing.T) {
	svc := &fakeDentistService{deleteFn: func(ctx context.Context, id string) error { return nil }}
	handler := httpadapter.NewDentistHandler(svc)
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/dentists/dentist-1", nil)
	req.SetPathValue("id", "dentist-1")
	rec := httptest.NewRecorder()

	handler.Delete(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)
}
