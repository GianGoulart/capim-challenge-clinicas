package integration_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	httpadapter "github.com/giancarlogoulart/capim-challenge-clinicas/internal/adapters/http"
	"github.com/giancarlogoulart/capim-challenge-clinicas/internal/adapters/memory"
	"github.com/giancarlogoulart/capim-challenge-clinicas/internal/adapters/pix"
	clinicapp "github.com/giancarlogoulart/capim-challenge-clinicas/internal/application/clinic"
	dentistapp "github.com/giancarlogoulart/capim-challenge-clinicas/internal/application/dentist"
	paymentapp "github.com/giancarlogoulart/capim-challenge-clinicas/internal/application/payment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestServer monta toda a stack da aplicação — adapters in-memory reais,
// use cases reais, router HTTP real — conectados exatamente como em
// cmd/api/main.go, mas com um simulador de Pix rápido para os testes não
// precisarem esperar vários segundos pela confirmação.
func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	clinicRepo := memory.NewClinicRepository()
	dentistRepo := memory.NewDentistRepository()
	paymentRepo := memory.NewPaymentRepository()
	fastPix := pix.NewSimulator(5*time.Millisecond, 10*time.Millisecond)

	clinicService := clinicapp.NewService(clinicRepo, dentistRepo)
	dentistService := dentistapp.NewService(dentistRepo, clinicRepo)
	paymentService := paymentapp.NewService(paymentRepo, clinicRepo, dentistRepo, fastPix)

	router := httpadapter.NewRouter(
		httpadapter.NewClinicHandler(clinicService),
		httpadapter.NewDentistHandler(dentistService),
		httpadapter.NewPaymentHandler(paymentService),
		"../../api/openapi.yaml",
	)

	return httptest.NewServer(router)
}

func postJSON(t *testing.T, url string, body map[string]any) *http.Response {
	t.Helper()
	b, err := json.Marshal(body)
	require.NoError(t, err)
	resp, err := http.Post(url, "application/json", bytes.NewReader(b))
	require.NoError(t, err)
	return resp
}

func decodeJSON(t *testing.T, resp *http.Response) map[string]any {
	t.Helper()
	defer resp.Body.Close()
	var out map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&out))
	return out
}

func TestClinicLifecycle_CreateGetUpdateDelete(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	createResp := postJSON(t, srv.URL+"/api/v1/clinics", map[string]any{
		"document": "52998224725", "corporate_name": "Clinica Sorriso LTDA",
		"trade_name": "Clinica Sorriso", "bank_code": "341", "agency": "1234", "account": "56789-0",
	})
	assert.Equal(t, http.StatusCreated, createResp.StatusCode)
	created := decodeJSON(t, createResp)
	id := created["id"].(string)
	require.NotEmpty(t, id)

	getResp, err := http.Get(srv.URL + "/api/v1/clinics/" + id)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, getResp.StatusCode)
	getResp.Body.Close()

	updateReq, err := http.NewRequest(http.MethodPut, srv.URL+"/api/v1/clinics/"+id,
		bytes.NewBufferString(`{"corporate_name":"New Corp","trade_name":"New Trade"}`))
	require.NoError(t, err)
	updateResp, err := http.DefaultClient.Do(updateReq)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, updateResp.StatusCode)
	updated := decodeJSON(t, updateResp)
	assert.Equal(t, "New Corp", updated["corporate_name"])

	delReq, err := http.NewRequest(http.MethodDelete, srv.URL+"/api/v1/clinics/"+id, nil)
	require.NoError(t, err)
	delResp, err := http.DefaultClient.Do(delReq)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, delResp.StatusCode)

	finalGet, err := http.Get(srv.URL + "/api/v1/clinics/" + id)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, finalGet.StatusCode)
}

func TestClinicCreate_InvalidDocumentReturns422(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	resp := postJSON(t, srv.URL+"/api/v1/clinics", map[string]any{
		"document": "123", "corporate_name": "Corp", "trade_name": "Trade",
		"bank_code": "341", "agency": "1", "account": "2",
	})

	assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
}

func TestDentistLifecycle_CreateUnderClinicAndList(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	clinicResp := postJSON(t, srv.URL+"/api/v1/clinics", map[string]any{
		"document": "52998224725", "corporate_name": "Corp", "trade_name": "Trade",
		"bank_code": "341", "agency": "1", "account": "2",
	})
	clinicID := decodeJSON(t, clinicResp)["id"].(string)

	dentistResp := postJSON(t, srv.URL+"/api/v1/clinics/"+clinicID+"/dentists", map[string]any{
		"name": "Dra. Ana", "phone": "+55 11 90000-0000", "email": "ana@example.com", "is_admin": true,
	})
	assert.Equal(t, http.StatusCreated, dentistResp.StatusCode)

	listResp, err := http.Get(srv.URL + "/api/v1/clinics/" + clinicID + "/dentists")
	require.NoError(t, err)
	defer listResp.Body.Close()
	var list []map[string]any
	require.NoError(t, json.NewDecoder(listResp.Body).Decode(&list))
	assert.Len(t, list, 1)
}

func TestClinicDelete_ConflictWhenDentistsLinkedReturns409(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	clinicResp := postJSON(t, srv.URL+"/api/v1/clinics", map[string]any{
		"document": "52998224725", "corporate_name": "Corp", "trade_name": "Trade",
		"bank_code": "341", "agency": "1", "account": "2",
	})
	clinicID := decodeJSON(t, clinicResp)["id"].(string)

	dentistResp := postJSON(t, srv.URL+"/api/v1/clinics/"+clinicID+"/dentists", map[string]any{
		"name": "Dra. Ana", "phone": "+55 11 90000-0000", "email": "ana@example.com", "is_admin": true,
	})
	require.Equal(t, http.StatusCreated, dentistResp.StatusCode)

	delReq, err := http.NewRequest(http.MethodDelete, srv.URL+"/api/v1/clinics/"+clinicID, nil)
	require.NoError(t, err)
	delResp, err := http.DefaultClient.Do(delReq)
	require.NoError(t, err)
	assert.Equal(t, http.StatusConflict, delResp.StatusCode)

	getResp, err := http.Get(srv.URL + "/api/v1/clinics/" + clinicID)
	require.NoError(t, err)
	defer getResp.Body.Close()
	assert.Equal(t, http.StatusOK, getResp.StatusCode, "clinic must not have been deleted")
}

func TestDentistCreate_UnknownClinicReturns404(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	resp := postJSON(t, srv.URL+"/api/v1/clinics/does-not-exist/dentists", map[string]any{
		"name": "Dra. Ana", "phone": "1", "email": "ana@example.com",
	})

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestPaymentLifecycle_CreatedThenAutoApprovedAsynchronously(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	clinicResp := postJSON(t, srv.URL+"/api/v1/clinics", map[string]any{
		"document": "52998224725", "corporate_name": "Corp", "trade_name": "Trade",
		"bank_code": "341", "agency": "1", "account": "2",
	})
	clinicID := decodeJSON(t, clinicResp)["id"].(string)

	payResp := postJSON(t, srv.URL+"/api/v1/payments", map[string]any{
		"clinic_id": clinicID, "amount_cents": 15000,
	})
	require.Equal(t, http.StatusCreated, payResp.StatusCode)
	created := decodeJSON(t, payResp)
	assert.Equal(t, "pending", created["status"])
	assert.NotEmpty(t, created["pix_code"])
	paymentID := created["id"].(string)

	assert.Eventually(t, func() bool {
		// o testify executa essa função de condição na sua própria goroutine
		// a cada tick; t.FailNow() (usado pelo require) não é seguro nesse
		// contexto, então usamos assert + retorno antecipado de false em vez
		// de require em qualquer falha.
		resp, err := http.Get(srv.URL + "/api/v1/payments/" + paymentID)
		if !assert.NoError(t, err) {
			return false
		}
		defer resp.Body.Close()
		var body map[string]any
		if !assert.NoError(t, json.NewDecoder(resp.Body).Decode(&body)) {
			return false
		}
		return body["status"] == "approved"
	}, 2*time.Second, 10*time.Millisecond, "payment should be auto-approved by the simulated Pix provider")
}

func TestPaymentCreate_UnknownDentistReturns404(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	clinicResp := postJSON(t, srv.URL+"/api/v1/clinics", map[string]any{
		"document": "52998224725", "corporate_name": "Corp", "trade_name": "Trade",
		"bank_code": "341", "agency": "1", "account": "2",
	})
	clinicID := decodeJSON(t, clinicResp)["id"].(string)

	resp := postJSON(t, srv.URL+"/api/v1/payments", map[string]any{
		"clinic_id": clinicID, "dentist_id": "does-not-exist", "amount_cents": 1000,
	})

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestPaymentCreate_DentistFromDifferentClinicReturns422(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	clinicAResp := postJSON(t, srv.URL+"/api/v1/clinics", map[string]any{
		"document": "52998224725", "corporate_name": "Corp A", "trade_name": "Trade A",
		"bank_code": "341", "agency": "1", "account": "2",
	})
	clinicAID := decodeJSON(t, clinicAResp)["id"].(string)

	clinicBResp := postJSON(t, srv.URL+"/api/v1/clinics", map[string]any{
		"document": "11144477735", "corporate_name": "Corp B", "trade_name": "Trade B",
		"bank_code": "341", "agency": "1", "account": "3",
	})
	clinicBID := decodeJSON(t, clinicBResp)["id"].(string)

	dentistResp := postJSON(t, srv.URL+"/api/v1/clinics/"+clinicBID+"/dentists", map[string]any{
		"name": "Dra. Ana", "phone": "+55 11 90000-0000", "email": "ana@example.com", "is_admin": true,
	})
	dentistID := decodeJSON(t, dentistResp)["id"].(string)

	resp := postJSON(t, srv.URL+"/api/v1/payments", map[string]any{
		"clinic_id": clinicAID, "dentist_id": dentistID, "amount_cents": 1000,
	})

	assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
}

func TestDocsAndOpenAPIEndpoints_AreServed(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	docsResp, err := http.Get(srv.URL + "/docs")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, docsResp.StatusCode)
	docsResp.Body.Close()

	specResp, err := http.Get(srv.URL + "/openapi.yaml")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, specResp.StatusCode)
	specResp.Body.Close()
}
