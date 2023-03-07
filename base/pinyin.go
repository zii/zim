package base

import (
	"strings"

	"github.com/mozillazg/go-pinyin"
)

func PinyinString(cn string) string {
	arg := pinyin.NewArgs()
	arg.Fallback = func(r rune, a pinyin.Args) []string {
		s := string(r)
		s = strings.ToLower(s)
		return []string{s}
	}
	ss := pinyin.LazyConvert(cn, &arg)
	return strings.Join(ss, "")
}

// 首字母
func PinyinInitials(cn string) string {
	s := PinyinString(cn)
	if len(s) > 0 {
		return s[:1]
	}
	return ""
}
