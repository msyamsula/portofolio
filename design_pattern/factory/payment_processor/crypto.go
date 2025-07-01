package paymentprocessor

import (
	"fmt"
)

type CryptoProcessor struct {
}

type CryptoMetadata struct {
	token string
}

func (p *CryptoProcessor) ProcessPayment(amount int, currency string, metadata ...any) bool {
	fmt.Println("processing payment in crypto", amount, currency, metadata)
	return true
}
