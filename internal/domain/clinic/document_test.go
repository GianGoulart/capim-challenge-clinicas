package clinic_test

import (
	"testing"

	"github.com/giancarlogoulart/capim-challenge-clinicas/internal/domain/clinic"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDocument_ValidCPF(t *testing.T) {
	doc, err := clinic.NewDocument("529.982.247-25")

	require.NoError(t, err)
	assert.Equal(t, clinic.DocumentTypeCPF, doc.Type())
	assert.Equal(t, "52998224725", doc.Digits())
}

func TestNewDocument_ValidCNPJ(t *testing.T) {
	doc, err := clinic.NewDocument("11.222.333/0001-81")

	require.NoError(t, err)
	assert.Equal(t, clinic.DocumentTypeCNPJ, doc.Type())
	assert.Equal(t, "11222333000181", doc.Digits())
}

func TestNewDocument_InvalidCPFCheckDigit(t *testing.T) {
	_, err := clinic.NewDocument("529.982.247-20")
	assert.Error(t, err)
}

func TestNewDocument_InvalidCNPJCheckDigit(t *testing.T) {
	_, err := clinic.NewDocument("11.222.333/0001-80")
	assert.Error(t, err)
}

func TestNewDocument_AllSameDigitsRejected(t *testing.T) {
	_, err := clinic.NewDocument("111.111.111-11")
	assert.Error(t, err)
}

func TestNewDocument_InvalidLength(t *testing.T) {
	_, err := clinic.NewDocument("123")
	assert.Error(t, err)
}
