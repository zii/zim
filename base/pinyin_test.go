package base

import (
	"fmt"
	"sort"
	"testing"
)

func TestPinyin(t *testing.T) {
	s := PinyinString("3暗黑破坏s的")
	fmt.Println(s)

	a := []string{"aaa", "AAA", "1"}
	sort.Strings(a)
	fmt.Println("sorted:", a)
}
