package payment_test

import (
	"testing"

	"github.com/giancarlogoulart/capim-challenge-clinicas/internal/domain/payment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMoney_Valid(t *testing.T) {
	m, err := payment.NewMoney(1050)

	require.NoError(t, err)
	assert.Equal(t, int64(1050), m.Cents())
	assert.Equal(t, "R$ 10.50", m.String())
}

func TestNewMoney_RejectsZero(t *testing.T) {
	_, err := payment.NewMoney(0)
	assert.Error(t, err)
}

func TestNewMoney_RejectsNegative(t *testing.T) {
	_, err := payment.NewMoney(-100)
	assert.Error(t, err)
}
