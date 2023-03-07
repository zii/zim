package main

import (
	"fmt"
	"strings"
)

const Indent = "    "

func ValueTToMd(vt any) string {
	switch a := vt.(type) {
	case string:
		return a
	case *TLType:
		return TypeToMd(a)
	}
	return fmt.Sprintf("%s", vt)
}

func FieldToMd(f *TLField, level int) string {
	out := f.Key + ": "
	switch a := f.ValueT.(type) {
	case string:
		out += a
		if f.Comment != "" {
			out += ` // ` + f.Comment
		}
	case *TLType:
		if f.Comment != "" {
			out = `// ` + f.Comment + "\n" + indents(level+1) + out
		}
		a.level = level + 1
		out += TypeToMd(a)
	}
	return out
}

func indents(n int) string {
	var out string
	for i := 0; i < n; i++ {
		out += Indent
	}
	return out
}

func TypeToMd(t *TLType) string {
	if t == nil {
		return ""
	}
	var out string
	if len(t.Fields) > 0 {
		out += "{\n"
		for _, f := range t.Fields {
			out += indents(t.level+1) + FieldToMd(f, t.level) + "\n"
		}
		out += indents(t.level) + "}"
		if t.IsArray {
			out = "[" + out + "]"
		}
	} else if t.Comment != "" {
		out += t.Comment
	}
	return out
}

func FunctionToMd(tf *TLFunction) string {
	if tf == nil {
		return ""
	}
	out := ""
	out += "------\n"
	out += fmt.Sprintf("### %s\n", tf.Title)
	out += fmt.Sprintf("#### %s\n", tf.URI())
	out += "##### 入参\n"
	out += "```\n"
	for _, p := range tf.Params {
		out += FieldToMd(p, 0) + "\n"
	}
	if len(tf.Params) == 0 {
		out += "无\n"
	}
	out += "```\n"
	out += "##### 出参\n"
	out += "```json\n"
	out += fmt.Sprintf("%s\n", TypeToMd(tf.Response))
	out += "```\n"
	return out
}

func toMdAnchor(s string) string {
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, "/", "")
	s = strings.ReplaceAll(s, "(", "")
	s = strings.ReplaceAll(s, ")", "")
	s = strings.ToLower(s)
	return s
}

func DocToMd(doc *TLDoc) string {
	out := ""
	out += fmt.Sprintf("# %s\n", doc.Title)
	out += fmt.Sprintf("## 目录\n")
	for _, f := range doc.Functions {
		out += fmt.Sprintf("- [%s](#%s)\n", f.Title, toMdAnchor(f.Title))
	}
	out += fmt.Sprintf("## 基本设置\n")
	out += "```\n"
	out += fmt.Sprintf("%s\n", doc.Intro)
	out += "```\n"
	for _, f := range doc.Functions {
		out += fmt.Sprintf("%s\n", FunctionToMd(f))
	}
	return out
}
