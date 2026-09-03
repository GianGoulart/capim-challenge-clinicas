package http

import (
	"time"

	clinicdomain "github.com/giancarlogoulart/capim-challenge-clinicas/internal/domain/clinic"
)

type clinicRequest struct {
	Document      string `json:"document"`
	CorporateName string `json:"corporate_name"`
	TradeName     string `json:"trade_name"`
	BankCode      string `json:"bank_code"`
	Agency        string `json:"agency"`
	Account       string `json:"account"`
}

type clinicUpdateRequest struct {
	CorporateName string `json:"corporate_name"`
	TradeName     string `json:"trade_name"`
}

type bankAccountRequest struct {
	BankCode string `json:"bank_code"`
	Agency   string `json:"agency"`
	Account  string `json:"account"`
}

type clinicResponse struct {
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

func toClinicResponse(c *clinicdomain.Clinic) clinicResponse {
	return clinicResponse{
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

func toClinicResponseList(clinics []*clinicdomain.Clinic) []clinicResponse {
	result := make([]clinicResponse, 0, len(clinics))
	for _, c := range clinics {
		result = append(result, toClinicResponse(c))
	}
	return result
}
