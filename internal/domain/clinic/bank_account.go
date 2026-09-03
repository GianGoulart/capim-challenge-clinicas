package clinic

import (
	"errors"
	"strings"
)

// BankAccount is a Value Object holding the clinic's banking details.
type BankAccount struct {
	BankCode string
	Agency   string
	Account  string
}

func NewBankAccount(bankCode, agency, account string) (BankAccount, error) {
	if strings.TrimSpace(bankCode) == "" {
		return BankAccount{}, errors.New("bank code is required")
	}
	if strings.TrimSpace(agency) == "" {
		return BankAccount{}, errors.New("agency is required")
	}
	if strings.TrimSpace(account) == "" {
		return BankAccount{}, errors.New("account is required")
	}
	return BankAccount{BankCode: bankCode, Agency: agency, Account: account}, nil
}
