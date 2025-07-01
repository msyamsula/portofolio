package paymentprocessor

type BlankProcessor struct {
}

func (p *BlankProcessor) ProcessPayment(amount int, currency string, metadata ...any) bool {
	return true
}
