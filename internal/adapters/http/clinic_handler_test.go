package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	httpadapter "github.com/giancarlogoulart/capim-challenge-clinicas/internal/adapters/http"
	clinicapp "github.com/giancarlogoulart/capim-challenge-clinicas/internal/application/clinic"
	clinicdomain "github.com/giancarlogoulart/capim-challenge-clinicas/internal/domain/clinic"
	"github.com/giancarlogoulart/capim-challenge-clinicas/internal/platform/apperrors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeClinicService struct {
	createFn func(ctx context.Context, input clinicapp.CreateInput) (*clinicdomain.Clinic, error)
	getFn    func(ctx context.Context, id string) (*clinicdomain.Clinic, error)
	listFn   func(ctx context.Context) ([]*clinicdomain.Clinic, error)
	updateFn func(ctx context.Context, id string, input clinicapp.UpdateInput) (*clinicdomain.Clinic, error)
	bankFn   func(ctx context.Context, id string, input clinicapp.UpdateBankAccountInput) (*clinicdomain.Clinic, error)
	deleteFn func(ctx context.Context, id string) error
}

func (f *fakeClinicService) Create(ctx context.Context, input clinicapp.CreateInput) (*clinicdomain.Clinic, error) {
	return f.createFn(ctx, input)
}
func (f *fakeClinicService) Get(ctx context.Context, id string) (*clinicdomain.Clinic, error) {
	return f.getFn(ctx, id)
}
func (f *fakeClinicService) List(ctx context.Context) ([]*clinicdomain.Clinic, error) {
	return f.listFn(ctx)
}
func (f *fakeClinicService) Update(ctx context.Context, id string, input clinicapp.UpdateInput) (*clinicdomain.Clinic, error) {
	return f.updateFn(ctx, id, input)
}
func (f *fakeClinicService) UpdateBankAccount(ctx context.Context, id string, input clinicapp.UpdateBankAccountInput) (*clinicdomain.Clinic, error) {
	return f.bankFn(ctx, id, input)
}
func (f *fakeClinicService) Delete(ctx context.Context, id string) error {
	return f.deleteFn(ctx, id)
}

func sampleClinic(t *testing.T) *clinicdomain.Clinic {
	t.Helper()
	doc, err := clinicdomain.NewDocument("52998224725")
	require.NoError(t, err)
	acc, err := clinicdomain.NewBankAccount("341", "1234", "56789-0")
	require.NoError(t, err)
	c, err := clinicdomain.NewClinic("id-1", doc, "Corp", "Trade", acc, timeNow())
	require.NoError(t, err)
	return c
}

func TestClinicHandler_Create_Success(t *testing.T) {
	svc := &fakeClinicService{
		createFn: func(ctx context.Context, input clinicapp.CreateInput) (*clinicdomain.Clinic, error) {
			return sampleClinic(t), nil
		},
	}
	handler := httpadapter.NewClinicHandler(svc)
	body := `{"document":"52998224725","corporate_name":"Corp","trade_name":"Trade","bank_code":"341","agency":"1234","account":"56789-0"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/clinics", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()

	handler.Create(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "id-1", resp["id"])
}

func TestClinicHandler_Create_InvalidBody(t *testing.T) {
	svc := &fakeClinicService{}
	handler := httpadapter.NewClinicHandler(svc)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/clinics", bytes.NewBufferString("not-json"))
	rec := httptest.NewRecorder()

	handler.Create(rec, req)

	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestClinicHandler_Create_ServiceError(t *testing.T) {
	svc := &fakeClinicService{
		createFn: func(ctx context.Context, input clinicapp.CreateInput) (*clinicdomain.Clinic, error) {
			return nil, apperrors.Validation("invalid document", nil)
		},
	}
	handler := httpadapter.NewClinicHandler(svc)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/clinics", bytes.NewBufferString(`{}`))
	rec := httptest.NewRecorder()

	handler.Create(rec, req)

	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestClinicHandler_Get_Success(t *testing.T) {
	svc := &fakeClinicService{
		getFn: func(ctx context.Context, id string) (*clinicdomain.Clinic, error) { return sampleClinic(t), nil },
	}
	handler := httpadapter.NewClinicHandler(svc)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/clinics/id-1", nil)
	req.SetPathValue("id", "id-1")
	rec := httptest.NewRecorder()

	handler.Get(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestClinicHandler_Get_NotFound(t *testing.T) {
	svc := &fakeClinicService{
		getFn: func(ctx context.Context, id string) (*clinicdomain.Clinic, error) {
			return nil, apperrors.NotFound("clinic not found")
		},
	}
	handler := httpadapter.NewClinicHandler(svc)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/clinics/missing", nil)
	req.SetPathValue("id", "missing")
	rec := httptest.NewRecorder()

	handler.Get(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestClinicHandler_List_Success(t *testing.T) {
	svc := &fakeClinicService{
		listFn: func(ctx context.Context) ([]*clinicdomain.Clinic, error) {
			return []*clinicdomain.Clinic{sampleClinic(t)}, nil
		},
	}
	handler := httpadapter.NewClinicHandler(svc)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/clinics", nil)
	rec := httptest.NewRecorder()

	handler.List(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp []map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Len(t, resp, 1)
}

func TestClinicHandler_Update_Success(t *testing.T) {
	svc := &fakeClinicService{
		updateFn: func(ctx context.Context, id string, input clinicapp.UpdateInput) (*clinicdomain.Clinic, error) {
			return sampleClinic(t), nil
		},
	}
	handler := httpadapter.NewClinicHandler(svc)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/clinics/id-1", bytes.NewBufferString(`{"corporate_name":"New","trade_name":"New"}`))
	req.SetPathValue("id", "id-1")
	rec := httptest.NewRecorder()

	handler.Update(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestClinicHandler_UpdateBankAccount_Success(t *testing.T) {
	svc := &fakeClinicService{
		bankFn: func(ctx context.Context, id string, input clinicapp.UpdateBankAccountInput) (*clinicdomain.Clinic, error) {
			return sampleClinic(t), nil
		},
	}
	handler := httpadapter.NewClinicHandler(svc)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/clinics/id-1/bank-account", bytes.NewBufferString(`{"bank_code":"001","agency":"1","account":"2"}`))
	req.SetPathValue("id", "id-1")
	rec := httptest.NewRecorder()

	handler.UpdateBankAccount(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestClinicHandler_Delete_Success(t *testing.T) {
	svc := &fakeClinicService{
		deleteFn: func(ctx context.Context, id string) error { return nil },
	}
	handler := httpadapter.NewClinicHandler(svc)
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/clinics/id-1", nil)
	req.SetPathValue("id", "id-1")
	rec := httptest.NewRecorder()

	handler.Delete(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestClinicHandler_Delete_NotFound(t *testing.T) {
	svc := &fakeClinicService{
		deleteFn: func(ctx context.Context, id string) error { return apperrors.NotFound("clinic not found") },
	}
	handler := httpadapter.NewClinicHandler(svc)
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/clinics/missing", nil)
	req.SetPathValue("id", "missing")
	rec := httptest.NewRecorder()

	handler.Delete(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}
