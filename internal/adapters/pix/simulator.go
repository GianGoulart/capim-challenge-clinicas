package pix

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/giancarlogoulart/capim-challenge-clinicas/internal/domain/payment"
)

// Simulator implementa payment.PixProvider sem nenhuma integração real com
// o Pix. Ele gera um código copia-e-cola falso e, após um delay aleatório
// dentro de [minDelay, maxDelay], invoca onApproved na sua própria
// goroutine — simulando uma confirmação de pagamento assíncrona.
type Simulator struct {
	minDelay time.Duration
	maxDelay time.Duration
}

// Asserção em tempo de compilação de que *Simulator satisfaz
// payment.PixProvider. Veja a seção "Compile-time interface assertions"
// no README.md para a justificativa.
var _ payment.PixProvider = (*Simulator)(nil)

func NewSimulator(minDelay, maxDelay time.Duration) *Simulator {
	return &Simulator{minDelay: minDelay, maxDelay: maxDelay}
}

// NewDefaultSimulator usa a janela de confirmação de 2-5 segundos sugerida
// pelo desafio.
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
