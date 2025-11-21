package paymentprocessor

import (
	"fmt"
)

type PaypalProcessor struct {
}

type PaypalMetadata struct {
	clientId, regionId string
}

func (p *PaypalProcessor) ProcessPayment(amount int, currency string, metadata ...any) bool {
	fmt.Println("processing payment in paypal", amount, currency, metadata)
	return true
}
