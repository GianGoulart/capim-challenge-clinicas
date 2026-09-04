// Package dto holds the HTTP wire-format types (request/response shapes) for
// every resource, plus the converters between them and domain entities.
//
// These types intentionally live outside the domain packages: they are the
// API contract (JSON field names, flattened Value Objects, optional fields),
// which changes for reasons unrelated to business rules (e.g. a front-end
// asking for a renamed field or extra pagination metadata) and must never
// leak transport concerns (json tags, wire-friendly primitives) into
// internal/domain. See the "Decisões de design notáveis" section in
// README.md for the full rationale.
package dto

import (
	"time"

	clinicdomain "github.com/giancarlogoulart/capim-challenge-clinicas/internal/domain/clinic"
)

// ClinicRequest is the request body for POST /clinics.
type ClinicRequest struct {
	Document      string `json:"document"`
	CorporateName string `json:"corporate_name"`
	TradeName     string `json:"trade_name"`
	BankCode      string `json:"bank_code"`
	Agency        string `json:"agency"`
	Account       string `json:"account"`
}

// ClinicUpdateRequest is the request body for PUT /clinics/{id}.
type ClinicUpdateRequest struct {
	CorporateName string `json:"corporate_name"`
	TradeName     string `json:"trade_name"`
}

// BankAccountRequest is the request body for PUT /clinics/{id}/bank-account.
type BankAccountRequest struct {
	BankCode string `json:"bank_code"`
	Agency   string `json:"agency"`
	Account  string `json:"account"`
}

// ClinicResponse is the response body for all clinic endpoints. It flattens
// the Document Value Object into plain string fields — internal/domain
// never appears in a JSON tag.
type ClinicResponse struct {
	ID            string    `json:"id"`
	Document      string    `json:"document"`
	DocumentType  string    `json:"document_type"`
	CorporateName string    `json:"corporate_name"`
	TradeName     string    `json:"trade_name"`
	BankCode      string    `json:"bank_code"`
	Agency        string    `json:"agency"`
	Account       string    `json:"account"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// ToClinicResponse converts a domain Clinic into its wire representation.
func ToClinicResponse(c *clinicdomain.Clinic) ClinicResponse {
	return ClinicResponse{
		ID:            c.ID,
		Document:      c.Document.Digits(),
		DocumentType:  string(c.Document.Type()),
		CorporateName: c.CorporateName,
		TradeName:     c.TradeName,
		BankCode:      c.BankAccount.BankCode,
		Agency:        c.BankAccount.Agency,
		Account:       c.BankAccount.Account,
		CreatedAt:     c.CreatedAt,
		UpdatedAt:     c.UpdatedAt,
	}
}

// ToClinicResponseList converts a slice of domain Clinics into their wire
// representation.
func ToClinicResponseList(clinics []*clinicdomain.Clinic) []ClinicResponse {
	result := make([]ClinicResponse, 0, len(clinics))
	for _, c := range clinics {
		result = append(result, ToClinicResponse(c))
	}
	return result
}
