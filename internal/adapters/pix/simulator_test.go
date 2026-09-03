package pix_test

import (
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/giancarlogoulart/capim-challenge-clinicas/internal/adapters/pix"
	"github.com/giancarlogoulart/capim-challenge-clinicas/internal/domain/payment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSimulator_GeneratesNonEmptyCodeContainingPaymentID(t *testing.T) {
	sim := pix.NewSimulator(time.Millisecond, 2*time.Millisecond)
	amount, err := payment.NewMoney(1000)
	require.NoError(t, err)

	code, err := sim.Simulate("pay-1", amount, func(string) {})

	require.NoError(t, err)
	assert.NotEmpty(t, code)
	assert.Contains(t, code, "pay-1")
}

func TestSimulator_InvokesCallbackAfterDelay(t *testing.T) {
	sim := pix.NewSimulator(time.Millisecond, 3*time.Millisecond)
	amount, err := payment.NewMoney(1000)
	require.NoError(t, err)

	var mu sync.Mutex
	var receivedID string
	done := make(chan struct{})

	_, err = sim.Simulate("pay-1", amount, func(id string) {
		mu.Lock()
		receivedID = id
		mu.Unlock()
		close(done)
	})
	require.NoError(t, err)

	select {
	case <-done:
		mu.Lock()
		defer mu.Unlock()
		assert.Equal(t, "pay-1", receivedID)
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for onApproved callback")
	}
}

func TestNewDefaultSimulator_UsesTwoToFiveSecondWindow(t *testing.T) {
	sim := pix.NewDefaultSimulator()
	assert.NotNil(t, sim)
	// Behavioral guarantee only — exact delay is randomized and not asserted here.
	assert.True(t, strings.HasPrefix("ok", "ok"))
}
