// Package dto contém os tipos de wire format HTTP (formatos de
// request/response) de cada recurso, além dos conversores entre eles e as
// entidades de domínio.
//
// Esses tipos vivem propositalmente fora dos pacotes de domínio: eles são o
// contrato da API (nomes de campo JSON, Value Objects "achatados", campos
// opcionais), que muda por motivos não relacionados a regras de negócio (ex:
// um front-end pedindo um campo renomeado ou metadados extras de paginação)
// e nunca deve deixar preocupações de transporte (tags json, primitivos
// amigáveis ao wire format) contaminarem internal/domain. Veja a seção
// "Decisões de design notáveis" no README.md para a justificativa completa.
package dto

import (
	"time"

	clinicdomain "github.com/giancarlogoulart/capim-challenge-clinicas/internal/domain/clinic"
)

// ClinicRequest é o corpo da requisição para POST /clinics.
type ClinicRequest struct {
	Document      string `json:"document"`
	CorporateName string `json:"corporate_name"`
	TradeName     string `json:"trade_name"`
	BankCode      string `json:"bank_code"`
	Agency        string `json:"agency"`
	Account       string `json:"account"`
}

// ClinicUpdateRequest é o corpo da requisição para PUT /clinics/{id}.
type ClinicUpdateRequest struct {
	CorporateName string `json:"corporate_name"`
	TradeName     string `json:"trade_name"`
}

// BankAccountRequest é o corpo da requisição para PUT /clinics/{id}/bank-account.
type BankAccountRequest struct {
	BankCode string `json:"bank_code"`
	Agency   string `json:"agency"`
	Account  string `json:"account"`
}

// ClinicResponse é o corpo da resposta de todos os endpoints de clínica. Ele
// "achata" o Value Object Document em campos string simples — internal/domain
// nunca aparece numa tag JSON.
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

// ToClinicResponse converte uma Clinic de domínio para sua representação de wire format.
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

// ToClinicResponseList converte um slice de Clinics de domínio para sua
// representação de wire format.
func ToClinicResponseList(clinics []*clinicdomain.Clinic) []ClinicResponse {
	result := make([]ClinicResponse, 0, len(clinics))
	for _, c := range clinics {
		result = append(result, ToClinicResponse(c))
	}
	return result
}
