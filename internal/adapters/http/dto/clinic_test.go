package dto_test

import (
	"testing"
	"time"

	"github.com/giancarlogoulart/capim-challenge-clinicas/internal/adapters/http/dto"
	clinicdomain "github.com/giancarlogoulart/capim-challenge-clinicas/internal/domain/clinic"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func sampleClinic(t *testing.T) *clinicdomain.Clinic {
	t.Helper()
	doc, err := clinicdomain.NewDocument("52998224725")
	require.NoError(t, err)
	acc, err := clinicdomain.NewBankAccount("341", "1234", "56789-0")
	require.NoError(t, err)
	c, err := clinicdomain.NewClinic("clinic-1", doc, "Clinica Sorriso LTDA", "Clinica Sorriso", acc, time.Now())
	require.NoError(t, err)
	return c
}

func TestToClinicResponse_FlattensDocumentValueObject(t *testing.T) {
	c := sampleClinic(t)

	resp := dto.ToClinicResponse(c)

	assert.Equal(t, "clinic-1", resp.ID)
	assert.Equal(t, "52998224725", resp.Document)
	assert.Equal(t, "CPF", resp.DocumentType)
	assert.Equal(t, "Clinica Sorriso LTDA", resp.CorporateName)
	assert.Equal(t, "341", resp.BankCode)
	assert.Equal(t, "1234", resp.Agency)
	assert.Equal(t, "56789-0", resp.Account)
}

func TestToClinicResponseList_ConvertsEachClinic(t *testing.T) {
	clinics := []*clinicdomain.Clinic{sampleClinic(t), sampleClinic(t)}

	result := dto.ToClinicResponseList(clinics)

	assert.Len(t, result, 2)
}

func TestToClinicResponseList_EmptyInputReturnsEmptySlice(t *testing.T) {
	result := dto.ToClinicResponseList(nil)

	assert.NotNil(t, result)
	assert.Empty(t, result)
}
