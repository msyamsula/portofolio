package factory

import (
	"testing"

	"github.com/msyamsula/portofolio/other_works/design_pattern/factory"
	paymentprocessor "github.com/msyamsula/portofolio/other_works/design_pattern/factory/payment_processor"
)

func TestFactory(t *testing.T) {
	// var processor *paypal.PaypalProcessor

	blank := factory.NewPaymentProcessor(factory.ProcessorTypeDefault)
	_, ok := blank.(*paymentprocessor.BlankProcessor)
	if !ok {
		t.Error("processor is not blank")
		return
	}

	paypal := factory.NewPaymentProcessor(factory.ProcessorTypePaypal)
	_, ok = paypal.(*paymentprocessor.PaypalProcessor)
	if !ok {
		t.Error("processor is not paypal")
		return
	}

	stripe := factory.NewPaymentProcessor(factory.ProcessorTypeStripe)
	_, ok = stripe.(*paymentprocessor.StripeProcessor)
	if !ok {
		t.Error("processor is not stripe")
		return
	}

	crypto := factory.NewPaymentProcessor(factory.ProcessorTypeCrypto)
	_, ok = crypto.(*paymentprocessor.CryptoProcessor)
	if !ok {
		t.Error("processor is not crypto")
		return
	}

}
