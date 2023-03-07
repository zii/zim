package main

import (
	"errors"
	"io/ioutil"
	"os"

	"zim.cn/base/log"

	urfave "github.com/urfave/cli/v2"
)

/*

本工具的作用是将tl格式的api文档转换为md格式的文档。
tl格式简介:
TL-Language 是 telegram 用来描述 MTProto 的一种自定义语言, 格式严谨且简洁, 很适合用来手写api原型.
TL (Type Language) serves to describe the used system of types, constructors, and existing functions.
In fact, the combinator description format presented in Binary Data Serialization is used.

用法:
tl md -o API.md API.txt

*/

var runCommand = &urfave.Command{
	Name:  "md",
	Usage: "转为md格式",
	Flags: []urfave.Flag{
		&urfave.StringFlag{
			Name:  "o",
			Usage: "",
			Value: "API.md",
		},
	},
	Action: func(cctx *urfave.Context) error {
		out := cctx.String("o")
		if cctx.NArg() != 1 {
			return errors.New("只能有一个参数")
		}
		in := cctx.Args().First()
		if in == "" {
			return errors.New("参数为空")
		}
		doc, err := ParseFile(in)
		if err != nil {
			return err
		}
		content := DocToMd(doc)
		err = ioutil.WriteFile(out, []byte(content), 0755)
		if err != nil {
			return err
		}
		return nil
	},
}

func main() {
	local := []*urfave.Command{
		runCommand,
	}

	app := &urfave.App{
		Name:                 "tl",
		Usage:                "tl md -o API.md API.txt",
		Version:              "1.0.1",
		EnableBashCompletion: true,
		Flags:                []urfave.Flag{},

		Commands: local,
	}
	app.Setup()

	if err := app.Run(os.Args); err != nil {
		log.Error("app.Run:", err)
		return
	}
}
