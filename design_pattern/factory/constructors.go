package factory

import paymentprocessor "github.com/msyamsula/portofolio/design_pattern/factory/payment_processor"

func newBlankProcessor() PaymentProcessor {
	return &paymentprocessor.BlankProcessor{}
}

func newCryptoProcessor() PaymentProcessor {
	return &paymentprocessor.CryptoProcessor{}
}

func newPaypal() PaymentProcessor {
	return &paymentprocessor.PaypalProcessor{}
}

func newStripeProcessor() PaymentProcessor {
	return &paymentprocessor.StripeProcessor{}
}
