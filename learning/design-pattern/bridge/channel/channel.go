package channel

type Channel interface {
	Send(m string) string
}
