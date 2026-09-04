package dto_test

import (
	"testing"
	"time"

	"github.com/giancarlogoulart/capim-challenge-clinicas/internal/adapters/http/dto"
	dentistdomain "github.com/giancarlogoulart/capim-challenge-clinicas/internal/domain/dentist"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func sampleDentist(t *testing.T) *dentistdomain.Dentist {
	t.Helper()
	d, err := dentistdomain.NewDentist("dentist-1", "clinic-1", "Dra. Ana", "+55 11 90000-0000", "ana@example.com", true, time.Now())
	require.NoError(t, err)
	return d
}

func TestToDentistResponse_MapsAllFields(t *testing.T) {
	d := sampleDentist(t)

	resp := dto.ToDentistResponse(d)

	assert.Equal(t, "dentist-1", resp.ID)
	assert.Equal(t, "clinic-1", resp.ClinicID)
	assert.Equal(t, "Dra. Ana", resp.Name)
	assert.Equal(t, "+55 11 90000-0000", resp.Phone)
	assert.Equal(t, "ana@example.com", resp.Email)
	assert.True(t, resp.IsAdmin)
}

func TestToDentistResponseList_ConvertsEachDentist(t *testing.T) {
	dentists := []*dentistdomain.Dentist{sampleDentist(t), sampleDentist(t)}

	result := dto.ToDentistResponseList(dentists)

	assert.Len(t, result, 2)
}

func TestToDentistResponseList_EmptyInputReturnsEmptySlice(t *testing.T) {
	result := dto.ToDentistResponseList(nil)

	assert.NotNil(t, result)
	assert.Empty(t, result)
}
