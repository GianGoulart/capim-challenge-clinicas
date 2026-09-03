package http_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	httpadapter "github.com/giancarlogoulart/capim-challenge-clinicas/internal/adapters/http"
	clinicdomain "github.com/giancarlogoulart/capim-challenge-clinicas/internal/domain/clinic"
	"github.com/stretchr/testify/assert"
)

func TestRouter_ListClinics_RoutesToHandler(t *testing.T) {
	clinicSvc := &fakeClinicService{
		listFn: func(ctx context.Context) ([]*clinicdomain.Clinic, error) { return nil, nil },
	}
	router := httpadapter.NewRouter(
		httpadapter.NewClinicHandler(clinicSvc),
		httpadapter.NewDentistHandler(&fakeDentistService{}),
		httpadapter.NewPaymentHandler(&fakePaymentService{}),
		"testdata/openapi.yaml",
	)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/clinics", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestRouter_UnknownRoute_Returns404(t *testing.T) {
	router := httpadapter.NewRouter(
		httpadapter.NewClinicHandler(&fakeClinicService{}),
		httpadapter.NewDentistHandler(&fakeDentistService{}),
		httpadapter.NewPaymentHandler(&fakePaymentService{}),
		"testdata/openapi.yaml",
	)

	req := httptest.NewRequest(http.MethodGet, "/unknown", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestApplyMiddleware_RecoversFromPanic(t *testing.T) {
	panicking := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	})
	wrapped := httpadapter.ApplyMiddleware(panicking)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	assert.NotPanics(t, func() { wrapped.ServeHTTP(rec, req) })
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}
