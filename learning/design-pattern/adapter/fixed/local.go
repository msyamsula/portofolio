package fixed

import "fmt"

type LocalLog struct{}

func (LocalLog) Log(m string) {
	fmt.Println(m)
}
