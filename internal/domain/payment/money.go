package payment

import (
	"errors"
	"fmt"
)

// Money representa um valor em centavos (inteiro), evitando problemas de
// arredondamento de ponto flutuante em valores monetários.
type Money struct {
	cents int64
}

func NewMoney(cents int64) (Money, error) {
	if cents <= 0 {
		return Money{}, errors.New("amount must be greater than zero")
	}
	return Money{cents: cents}, nil
}

func (m Money) Cents() int64 { return m.cents }

func (m Money) String() string {
	return fmt.Sprintf("R$ %d.%02d", m.cents/100, m.cents%100)
}
