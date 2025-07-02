package fixed

import "fmt"

type SysLog struct{}

func (SysLog) Log(level int, m string) {
	fmt.Printf("level: %d, message: %s\n", level, m)
}
