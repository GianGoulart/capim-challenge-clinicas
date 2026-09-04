package http

import "net/http"

// NewRouter conecta todos os handlers de recursos sob /api/v1, além de /docs e
// /openapi.yaml para a documentação da API. openapiPath é o caminho no sistema de arquivos
// para o contrato OpenAPI servido em /openapi.yaml.
func NewRouter(clinicHandler *ClinicHandler, dentistHandler *DentistHandler, paymentHandler *PaymentHandler, openapiPath string) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/v1/clinics", clinicHandler.Create)
	mux.HandleFunc("GET /api/v1/clinics", clinicHandler.List)
	mux.HandleFunc("GET /api/v1/clinics/{id}", clinicHandler.Get)
	mux.HandleFunc("PUT /api/v1/clinics/{id}", clinicHandler.Update)
	mux.HandleFunc("PUT /api/v1/clinics/{id}/bank-account", clinicHandler.UpdateBankAccount)
	mux.HandleFunc("DELETE /api/v1/clinics/{id}", clinicHandler.Delete)

	mux.HandleFunc("POST /api/v1/clinics/{clinic_id}/dentists", dentistHandler.Create)
	mux.HandleFunc("GET /api/v1/clinics/{clinic_id}/dentists", dentistHandler.ListByClinic)
	mux.HandleFunc("GET /api/v1/dentists/{id}", dentistHandler.Get)
	mux.HandleFunc("PUT /api/v1/dentists/{id}", dentistHandler.Update)
	mux.HandleFunc("DELETE /api/v1/dentists/{id}", dentistHandler.Delete)

	mux.HandleFunc("POST /api/v1/payments", paymentHandler.Create)
	mux.HandleFunc("GET /api/v1/payments/{id}", paymentHandler.Get)

	mux.HandleFunc("GET /docs", docsHandler)
	mux.HandleFunc("GET /openapi.yaml", openapiHandler(openapiPath))

	return ApplyMiddleware(mux)
}
