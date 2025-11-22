package handler

type handler struct {
	*httpHandler
	*sqsConsumer
}
