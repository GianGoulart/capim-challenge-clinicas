package clinic

import (
	"fmt"
	"regexp"
	"strconv"
)

type DocumentType string

const (
	DocumentTypeCPF  DocumentType = "CPF"
	DocumentTypeCNPJ DocumentType = "CNPJ"
)

var nonDigitRE = regexp.MustCompile(`\D`)

// Document é um Value Object que representa um CPF ou CNPJ brasileiro,
// validado pelo seu algoritmo oficial de dígito verificador (check-digit).
type Document struct {
	docType DocumentType
	digits  string
}

func (d Document) Type() DocumentType { return d.docType }
func (d Document) Digits() string     { return d.digits }
func (d Document) String() string     { return d.digits }

// NewDocument faz o parse de raw (com ou sem pontuação) e o valida como
// um CPF (11 dígitos) ou CNPJ (14 dígitos) usando o algoritmo oficial de
// dígito verificador (check-digit).
func NewDocument(raw string) (Document, error) {
	digits := nonDigitRE.ReplaceAllString(raw, "")

	switch len(digits) {
	case 11:
		if !isValidCPF(digits) {
			return Document{}, fmt.Errorf("invalid CPF check digits: %q", raw)
		}
		return Document{docType: DocumentTypeCPF, digits: digits}, nil
	case 14:
		if !isValidCNPJ(digits) {
			return Document{}, fmt.Errorf("invalid CNPJ check digits: %q", raw)
		}
		return Document{docType: DocumentTypeCNPJ, digits: digits}, nil
	default:
		return Document{}, fmt.Errorf("document must have 11 (CPF) or 14 (CNPJ) digits, got %d", len(digits))
	}
}

func isAllSameDigit(s string) bool {
	for i := 1; i < len(s); i++ {
		if s[i] != s[0] {
			return false
		}
	}
	return true
}

func checkDigit(digits string, weights []int) int {
	sum := 0
	for i, w := range weights {
		n, _ := strconv.Atoi(string(digits[i]))
		sum += n * w
	}
	rem := sum % 11
	if rem < 2 {
		return 0
	}
	return 11 - rem
}

func isValidCPF(d string) bool {
	if isAllSameDigit(d) {
		return false
	}
	dv1 := checkDigit(d[:9], []int{10, 9, 8, 7, 6, 5, 4, 3, 2})
	dv2 := checkDigit(d[:10], []int{11, 10, 9, 8, 7, 6, 5, 4, 3, 2})
	return d[9:] == fmt.Sprintf("%d%d", dv1, dv2)
}

func isValidCNPJ(d string) bool {
	if isAllSameDigit(d) {
		return false
	}
	dv1 := checkDigit(d[:12], []int{5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2})
	dv2 := checkDigit(d[:13], []int{6, 5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2})
	return d[12:] == fmt.Sprintf("%d%d", dv1, dv2)
}
