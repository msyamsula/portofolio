package fixed

import "fmt"

type FluentD struct{}

func (FluentD) Log(tag string, m string) {
	fmt.Printf("%s:%s\n", tag, m)
}
