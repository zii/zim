package main

import (
	"fmt"
	"io/ioutil"
	"sort"
	"strings"
	"unicode"

	"github.com/forPelevin/gomoji"
)

type TLField struct {
	Key     string
	ValueT  any // 一般是 string, 也有可能是*TLType
	Comment string
}

type TLType struct {
	Name    string
	Comment string
	Fields  []*TLField
	IsArray bool
	level   int
}

type TLFunction struct {
	Name     string
	Title    string
	Comment  string
	Params   []*TLField
	Response *TLType
}

func (f *TLFunction) Path() string {
	ps := strings.Split(f.Name, ".")
	out := ""
	for _, p := range ps {
		out += "/" + p
	}
	return out
}

func (f *TLFunction) URI() string {
	return fmt.Sprintf("POST %s", f.Path())
}

type TLDoc struct {
	Title     string
	Intro     string
	Types     []*TLType // {name: *TLType}
	Functions []*TLFunction
}

func ParseString(fstr string) (*TLDoc, error) {
	lines := strings.Split(fstr, "\n")
	//for i := range lines {
	//	lines[i] = strings.TrimSpace(lines[i])
	//}
	out := &TLDoc{}
	if len(lines) == 0 {
		return out, nil
	}
	out.Title = lines[0]
	lines = lines[1:]
	out.Intro, lines = readIntroSection(lines)
	out.Types, lines = readTypesSection(lines)
	out.Functions = parseFunctionsSection(lines, out.Types)
	return out, nil
}

// 返回intro部分,和剩余行
func readIntroSection(lines []string) (string, []string) {
	for i, l := range lines {
		if strings.HasPrefix(l, "---") && strings.Contains(l, "--- type ---") {
			intro := strings.Join(lines[:i], "\n")
			intro = strings.TrimSpace(intro)
			return intro, lines[i+1:]
		}
	}
	return "", nil
}

// 解析--- type ---部分
func readTypesSection(lines []string) ([]*TLType, []string) {
	var out []*TLType
	const (
		readingNothing = 0
		readingComment = 1
		readingType    = 2
	)
	var state int
	var tt = &TLType{}
	var typestr string

	for i, l := range lines {
		if strings.HasPrefix(l, "---") && strings.Contains(l, "--- function ---") {
			return out, lines[i+1:]
		}
		if l == "" {
			tt = &TLType{}
			state = readingNothing
			continue
		}
		if state == readingNothing && strings.HasPrefix(l, `//`) {
			state = readingComment
			tt.Comment = strings.TrimSpace(l[2:])
		} else if state == readingComment {
			if l[0] != '#' && strings.Index(l, "#") > 0 {
				state = readingType
				typestr = ""
			} else {
				tt.Comment += "\n" + l
			}
		}
		if state == readingType {
			typestr += " " + l
			if strings.HasSuffix(l, ";") {
				typestr = strings.TrimSpace(typestr)
				t := parseTypeLine(typestr, out)
				tt.Fields = t.Fields
				tt.Name = t.Name
				tt.Comment = fillFieldComment(tt.Comment, tt.Fields)
				state = readingNothing
				out = append(out, tt)
			}
		}
	}
	return out, nil
}

func parseTypeLine(line string, types []*TLType) *TLType {
	i := strings.Index(line, "#")
	if i < 0 {
		err := fmt.Errorf("parseTypeLine error, not found # charactor. near: %s", line)
		panic(err)
	}
	out := &TLType{}
	out.Name = line[:i]
	line = line[i+1:]
	line = strings.TrimSpace(line)
	line = strings.TrimRight(line, ";")
	tm := make(map[string]*TLType)
	for _, t := range types {
		tm[t.Name] = t
	}
	secs := strings.Split(line, " ")
	for _, sec := range secs {
		fs := strings.SplitN(sec, ":", 2)
		if len(fs) == 2 {
			vt := fs[1]
			f := &TLField{
				Key:    fs[0],
				ValueT: fs[1],
			}
			if vt != "" && unicode.IsUpper(rune(vt[0])) {
				if t, ok := tm[vt]; ok {
					f.ValueT = t
				}
			} else if vt != "" && vt[0] == '[' {
				vt = strings.TrimRight(vt[1:], "]")
				if t, ok := tm[vt]; ok {
					t2 := *t
					t2.IsArray = true
					f.ValueT = &t2
				}
			}
			out.Fields = append(out.Fields, f)
		}
	}
	return out
}

func findCommentKey(comment, key string) int {
	offset := 0
	for {
		i := strings.Index(comment, key)
		if i < 0 {
			return -1
		}
		if i > 0 && !unicode.IsSpace(rune(comment[i-1])) {
			comment = comment[i+len(key):]
			offset = i + len(key)
			continue
		}
		return i + offset
	}
}

// 从注释里提取每个字段的注释, 填充进TLField, 返回真正的注释
func fillFieldComment(comment string, fields []*TLField) string {
	var indexs []int
	fm := make(map[string]*TLField)
	for _, f := range fields {
		fm[f.Key] = f
		i := findCommentKey(comment, f.Key+":")
		if i > 0 {
			indexs = append(indexs, i)
		}
	}
	sort.Ints(indexs)
	n := len(indexs)
	for i, id := range indexs {
		var s string
		if i < n-1 {
			s = comment[id:indexs[i+1]]
		} else {
			s = comment[id:]
		}
		fs := strings.SplitN(s, ":", 2)
		if len(fs) == 2 {
			key := strings.TrimSpace(fs[0])
			val := strings.TrimSpace(fs[1])
			f := fm[key]
			if f != nil {
				f.Comment = val
			}
		}

	}
	primary_comment := comment
	if len(indexs) > 0 {
		primary_comment = strings.TrimSpace(comment[:indexs[0]])
	}
	return primary_comment
}

func commentToTitle(c string) string {
	fs := strings.SplitN(c, " ", 2)
	if len(fs) == 0 {
		return c
	}
	title := fs[0]
	title = gomoji.RemoveEmojis(title)
	return title
}

// 解析 --- function --- 部分
func parseFunctionsSection(lines []string, types []*TLType) []*TLFunction {
	var out []*TLFunction

	const (
		readingNothing  = 0
		readingComment  = 1
		readingFunction = 2
	)
	var state int
	var tf = &TLFunction{}
	var funcstr string
	for _, l := range lines {
		if l == "" {
			tf = &TLFunction{}
			state = readingNothing
			continue
		}
		if state == readingNothing && strings.HasPrefix(l, `#`) {
			state = readingComment
			tf.Comment = strings.TrimSpace(l[1:])
		} else if state == readingComment {
			if l[0] != '#' && strings.Index(l, ".") > 0 && strings.Index(l, "#") > 0 {
				state = readingFunction
				funcstr = ""
			} else {
				tf.Comment += "\n" + l
			}
		}
		if state == readingFunction {
			funcstr += "\n" + l
			if strings.HasSuffix(l, ";") {
				funcstr = strings.TrimSpace(funcstr)
				t := parseFuncLine(funcstr, types)
				tf.Params = t.Params
				tf.Name = t.Name
				tf.Comment = fillFieldComment(tf.Comment, tf.Params)
				tf.Title = commentToTitle(tf.Comment)
				tf.Response = t.Response
				state = readingNothing
				out = append(out, tf)
			}
		}
	}
	return out
}

func parseFuncLine(line string, types []*TLType) *TLFunction {
	i := strings.Index(line, "#")
	if i < 0 {
		err := fmt.Errorf("not found # charactor. near: %s", line)
		panic(err)
	}
	out := &TLFunction{}
	out.Name = line[:i]
	line = line[i+1:]
	i = strings.Index(line, "=")
	if i < 0 {
		err := fmt.Errorf("not found = charactor. near: %s", line)
		panic(err)
	}
	tm := make(map[string]*TLType)
	for _, t := range types {
		tm[t.Name] = t
	}
	paramstr := line[:i]
	secs := strings.Split(paramstr, " ")
	for _, sec := range secs {
		fs := strings.SplitN(sec, ":", 2)
		if len(fs) == 2 {
			vt := fs[1]
			f := &TLField{
				Key:    fs[0],
				ValueT: vt,
			}
			if vt != "" && unicode.IsUpper(rune(vt[0])) {
				if t, ok := tm[vt]; ok {
					f.ValueT = t
				}
			} else if vt != "" && vt[0] == '[' {
				vt = strings.TrimRight(vt[1:], "]")
				if t, ok := tm[vt]; ok {
					t2 := *t
					t2.IsArray = true
					f.ValueT = &t2
				}
			}
			out.Params = append(out.Params, f)
		}
	}
	respstr := strings.TrimSpace(line[i+1:])
	respstr = strings.TrimRight(respstr, ";")
	if strings.HasPrefix(respstr, "{") {
		out.Response = &TLType{
			Name:    "",
			Comment: respstr,
		}
	} else if strings.HasPrefix(respstr, "[") {
		inner := strings.TrimRight(respstr[1:], "]")
		p := tm[inner]
		if p != nil {
			t := *p
			t.IsArray = true
			out.Response = &t
		} else {
			out.Response = &TLType{
				Comment: respstr,
			}
		}
	} else {
		p := tm[respstr]
		if p != nil {
			t := *p
			out.Response = &t
		} else {
			out.Response = &TLType{
				Comment: respstr,
			}
		}
	}
	return out
}

func ParseFile(file string) (*TLDoc, error) {
	b, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	fstr := string(b)
	return ParseString(fstr)
}
