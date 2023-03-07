package uuid

import (
	"fmt"
	"testing"
	"time"
)

func Test2(t *testing.T) {
	MustInit("cat")
	st := time.Now()
	id := ""
	for i := 0; i < 100000; i++ {
		id = NextIDString("cat")
	}
	fmt.Println("took:", time.Since(st))
	fmt.Println(id)
}
