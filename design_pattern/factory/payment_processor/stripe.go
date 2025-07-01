package paymentprocessor

import (
	"fmt"
)

type StripeProcessor struct {
}

type StripeMetadata struct {
	jwt string
}

func (p *StripeProcessor) ProcessPayment(amount int, currency string, metadata ...any) bool {
	fmt.Println("processing payment in stripe", amount, currency, metadata)
	return true
}
