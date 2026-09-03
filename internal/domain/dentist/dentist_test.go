package dentist_test

import (
	"testing"
	"time"

	"github.com/giancarlogoulart/capim-challenge-clinicas/internal/domain/dentist"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDentist_Valid(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	d, err := dentist.NewDentist("id-1", "clinic-1", "Dra. Ana", "+55 11 90000-0000", "ana@example.com", true, now)

	require.NoError(t, err)
	assert.Equal(t, "id-1", d.ID)
	assert.Equal(t, "clinic-1", d.ClinicID)
	assert.True(t, d.IsAdmin)
	assert.Equal(t, now, d.CreatedAt)
}

func TestNewDentist_MissingClinicID(t *testing.T) {
	_, err := dentist.NewDentist("id-1", "", "Dra. Ana", "+55 11 90000-0000", "ana@example.com", false, time.Now())
	assert.Error(t, err)
}

func TestNewDentist_MissingName(t *testing.T) {
	_, err := dentist.NewDentist("id-1", "clinic-1", "", "+55 11 90000-0000", "ana@example.com", false, time.Now())
	assert.Error(t, err)
}

func TestNewDentist_MissingPhone(t *testing.T) {
	_, err := dentist.NewDentist("id-1", "clinic-1", "Dra. Ana", "", "ana@example.com", false, time.Now())
	assert.Error(t, err)
}

func TestNewDentist_InvalidEmail(t *testing.T) {
	_, err := dentist.NewDentist("id-1", "clinic-1", "Dra. Ana", "+55 11 90000-0000", "not-an-email", false, time.Now())
	assert.Error(t, err)
}

func TestDentist_Update(t *testing.T) {
	created := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	updated := created.Add(time.Hour)
	d, err := dentist.NewDentist("id-1", "clinic-1", "Dra. Ana", "+55 11 90000-0000", "ana@example.com", false, created)
	require.NoError(t, err)

	err = d.Update("Dra. Ana Silva", "+55 11 91111-1111", "ana.silva@example.com", true, updated)

	require.NoError(t, err)
	assert.Equal(t, "Dra. Ana Silva", d.Name)
	assert.True(t, d.IsAdmin)
	assert.Equal(t, updated, d.UpdatedAt)
}

func TestDentist_Update_RejectsInvalidEmail(t *testing.T) {
	d, err := dentist.NewDentist("id-1", "clinic-1", "Dra. Ana", "+55 11 90000-0000", "ana@example.com", false, time.Now())
	require.NoError(t, err)

	err = d.Update("Dra. Ana", "+55 11 90000-0000", "bad-email", false, time.Now())
	assert.Error(t, err)
}
