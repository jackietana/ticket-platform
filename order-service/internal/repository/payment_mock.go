package repository

import (
	"context"
	"errors"
	"math/rand"
	"time"

	"github.com/google/uuid"
)

type FakePaymentProvider struct{}

func NewFakePaymentProvider() *FakePaymentProvider {
	return &FakePaymentProvider{}
}

func (p *FakePaymentProvider) ProcessPayment(ctx context.Context, orderID uuid.UUID, amountCents int) (string, error) {
	select {
	case <-time.After(300 * time.Millisecond):
	case <-ctx.Done():
		return "", ctx.Err()
	}

	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	if rnd.Intn(100) < 15 {
		return "", errors.New("payment_rejected: insufficient funds or invalid card details")
	}

	paymentTransactionID := "tx_" + uuid.New().String()[:8]
	return paymentTransactionID, nil
}
