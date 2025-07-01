package factory

import "sync"

const (
	ProcessorTypeDefault int = iota
	ProcessorTypePaypal
	ProcessorTypeStripe
	ProcessorTypeCrypto
)

var registryLock sync.RWMutex
var registry = map[int]func() PaymentProcessor{
	ProcessorTypeDefault: newBlankProcessor,
	ProcessorTypeCrypto:  newCryptoProcessor,
	ProcessorTypeStripe:  newStripeProcessor,
	ProcessorTypePaypal:  newPaypal,
}

type PaymentProcessor interface {
	ProcessPayment(amount int, currency string, metadata ...any) bool
}

func NewPaymentProcessor(processorType int) PaymentProcessor {
	if constructor, ok := registry[processorType]; ok {
		return constructor()
	}

	return registry[ProcessorTypeDefault]()
}

func Register(processorType int, p func() PaymentProcessor) {
	registryLock.Lock()
	defer registryLock.Unlock()
	registry[processorType] = p
}
