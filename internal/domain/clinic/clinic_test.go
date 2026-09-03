package clinic_test

import (
	"testing"
	"time"

	"github.com/giancarlogoulart/capim-challenge-clinicas/internal/domain/clinic"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func validDocument(t *testing.T) clinic.Document {
	t.Helper()
	doc, err := clinic.NewDocument("52998224725")
	require.NoError(t, err)
	return doc
}

func validBankAccount(t *testing.T) clinic.BankAccount {
	t.Helper()
	acc, err := clinic.NewBankAccount("341", "1234", "56789-0")
	require.NoError(t, err)
	return acc
}

func TestNewClinic_Valid(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	c, err := clinic.NewClinic("id-1", validDocument(t), "Clinica Sorriso LTDA", "Clinica Sorriso", validBankAccount(t), now)

	require.NoError(t, err)
	assert.Equal(t, "id-1", c.ID)
	assert.Equal(t, "Clinica Sorriso LTDA", c.CorporateName)
	assert.Equal(t, "Clinica Sorriso", c.TradeName)
	assert.Equal(t, now, c.CreatedAt)
	assert.Equal(t, now, c.UpdatedAt)
}

func TestNewClinic_MissingCorporateName(t *testing.T) {
	_, err := clinic.NewClinic("id-1", validDocument(t), "", "Clinica Sorriso", validBankAccount(t), time.Now())
	assert.Error(t, err)
}

func TestNewClinic_MissingTradeName(t *testing.T) {
	_, err := clinic.NewClinic("id-1", validDocument(t), "Clinica Sorriso LTDA", "", validBankAccount(t), time.Now())
	assert.Error(t, err)
}

func TestClinic_UpdateInfo(t *testing.T) {
	created := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	updated := created.Add(24 * time.Hour)
	c, err := clinic.NewClinic("id-1", validDocument(t), "Old Corp", "Old Trade", validBankAccount(t), created)
	require.NoError(t, err)

	err = c.UpdateInfo("New Corp", "New Trade", updated)

	require.NoError(t, err)
	assert.Equal(t, "New Corp", c.CorporateName)
	assert.Equal(t, "New Trade", c.TradeName)
	assert.Equal(t, updated, c.UpdatedAt)
}

func TestClinic_UpdateInfo_RejectsEmptyCorporateName(t *testing.T) {
	c, err := clinic.NewClinic("id-1", validDocument(t), "Old Corp", "Old Trade", validBankAccount(t), time.Now())
	require.NoError(t, err)

	err = c.UpdateInfo("", "New Trade", time.Now())
	assert.Error(t, err)
}

func TestClinic_UpdateBankAccount(t *testing.T) {
	created := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	updated := created.Add(24 * time.Hour)
	c, err := clinic.NewClinic("id-1", validDocument(t), "Corp", "Trade", validBankAccount(t), created)
	require.NoError(t, err)

	newAccount, err := clinic.NewBankAccount("001", "0001", "111-1")
	require.NoError(t, err)

	c.UpdateBankAccount(newAccount, updated)

	assert.Equal(t, newAccount, c.BankAccount)
	assert.Equal(t, updated, c.UpdatedAt)
}
