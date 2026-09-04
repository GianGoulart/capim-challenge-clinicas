package clinic

import (
	"errors"
	"strings"
	"time"
)

// Clinic é o aggregate root de uma clínica odontológica e seus dados bancários.
type Clinic struct {
	ID            string
	Document      Document
	CorporateName string
	TradeName     string
	BankAccount   BankAccount
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func NewClinic(id string, document Document, corporateName, tradeName string, bankAccount BankAccount, now time.Time) (*Clinic, error) {
	if strings.TrimSpace(id) == "" {
		return nil, errors.New("id is required")
	}
	if strings.TrimSpace(corporateName) == "" {
		return nil, errors.New("corporate name is required")
	}
	if strings.TrimSpace(tradeName) == "" {
		return nil, errors.New("trade name is required")
	}
	return &Clinic{
		ID:            id,
		Document:      document,
		CorporateName: corporateName,
		TradeName:     tradeName,
		BankAccount:   bankAccount,
		CreatedAt:     now,
		UpdatedAt:     now,
	}, nil
}

func (c *Clinic) UpdateInfo(corporateName, tradeName string, now time.Time) error {
	if strings.TrimSpace(corporateName) == "" {
		return errors.New("corporate name is required")
	}
	if strings.TrimSpace(tradeName) == "" {
		return errors.New("trade name is required")
	}
	c.CorporateName = corporateName
	c.TradeName = tradeName
	c.UpdatedAt = now
	return nil
}

func (c *Clinic) UpdateBankAccount(bankAccount BankAccount, now time.Time) {
	c.BankAccount = bankAccount
	c.UpdatedAt = now
}
