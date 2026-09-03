package clinic_test

import (
	"testing"

	"github.com/giancarlogoulart/capim-challenge-clinicas/internal/domain/clinic"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBankAccount_Valid(t *testing.T) {
	acc, err := clinic.NewBankAccount("341", "1234", "56789-0")

	require.NoError(t, err)
	assert.Equal(t, "341", acc.BankCode)
	assert.Equal(t, "1234", acc.Agency)
	assert.Equal(t, "56789-0", acc.Account)
}

func TestNewBankAccount_MissingBankCode(t *testing.T) {
	_, err := clinic.NewBankAccount("", "1234", "56789-0")
	assert.Error(t, err)
}

func TestNewBankAccount_MissingAgency(t *testing.T) {
	_, err := clinic.NewBankAccount("341", "", "56789-0")
	assert.Error(t, err)
}

func TestNewBankAccount_MissingAccount(t *testing.T) {
	_, err := clinic.NewBankAccount("341", "1234", "")
	assert.Error(t, err)
}
