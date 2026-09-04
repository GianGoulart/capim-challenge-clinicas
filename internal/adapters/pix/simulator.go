package pix

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/giancarlogoulart/capim-challenge-clinicas/internal/domain/payment"
)

// Simulator implements payment.PixProvider without any real Pix
// integration. It generates a fake copy-and-paste code and, after a random
// delay within [minDelay, maxDelay], invokes onApproved on its own
// goroutine — simulating an asynchronous payment confirmation.
type Simulator struct {
	minDelay time.Duration
	maxDelay time.Duration
}

// Compile-time assertion that *Simulator satisfies payment.PixProvider.
// See the "Compile-time interface assertions" section in README.md for
// the rationale.
var _ payment.PixProvider = (*Simulator)(nil)

func NewSimulator(minDelay, maxDelay time.Duration) *Simulator {
	return &Simulator{minDelay: minDelay, maxDelay: maxDelay}
}

// NewDefaultSimulator uses the challenge's suggested 2-5 second
// confirmation window.
func NewDefaultSimulator() *Simulator {
	return NewSimulator(2*time.Second, 5*time.Second)
}

func (s *Simulator) Simulate(paymentID string, amount payment.Money, onApproved func(paymentID string)) (string, error) {
	pixCode := fmt.Sprintf("00020126SIMULATEDPIX%s5204000053039865406%s5802BR6304%s",
		paymentID, amount.String(), paymentID)

	delay := s.minDelay
	if window := s.maxDelay - s.minDelay; window > 0 {
		delay += time.Duration(rand.Int63n(int64(window)))
	}

	go func() {
		time.Sleep(delay)
		onApproved(paymentID)
	}()

	return pixCode, nil
}
