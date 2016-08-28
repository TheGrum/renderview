// Copyright 2016 Howard C. Shaw III. All rights reserved.
// Use of this source code is governed by the MIT-license
// as defined in the LICENSE file.

package main

import (
	"bytes"
	"flag"
	"html/template"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/TheGrum/renderview/driver"

	rv "github.com/TheGrum/renderview"
)

var (
	defaultFlags = flag.Bool("defaultflags", true, "include default flags (left, top, right, bottom, width, height, options)")
	extraFlags   = flag.String("extraflags", "", "quoted comma-separated list, name, type, and starting value, e.g. \"left,float64,0,top,float64,0\"")
	watchFile    = flag.String("watch", "", "if defined, image file to read in after executing command")
)

func main() {
	flag.Parse()

	cmd := flag.Arg(0)
	args := flag.Arg(1)
	argTemplate := template.Must(template.New("arg").Parse(args))

	m := rv.NewBasicRenderModel()
	if *defaultFlags {
		m.AddParameters(
			rv.SetHints(rv.HINT_SIDEBAR,
				rv.NewFloat64RP("left", 0),
				rv.NewFloat64RP("top", 0),
				rv.NewFloat64RP("right", 100),
				rv.NewFloat64RP("bottom", 100),
				rv.NewIntRP("width", 100),
				rv.NewIntRP("height", 100),
				rv.NewIntRP("options", rv.OPT_AUTO_ZOOM))...)
	}
	m.AddParameters(rv.SetHints(rv.HINT_SIDEBAR, createExtraFlags(*extraFlags)...)...)
	m.InnerRender = getInnerRender(m, cmd, argTemplate)
	m.InnerRender()
	driver.Main(m)
}

func getInnerRender(m *rv.BasicRenderModel, cmd string, argtemplate *template.Template) func() {
	return func() {
		InnerRender(m, cmd, argtemplate)
	}
}

func InnerRender(m *rv.BasicRenderModel, cmd string, argtemplate *template.Template) {
	flags := m.GetParameterNames()
	templateMap := make(map[string]string)
	for _, k := range flags {
		p := m.GetParameter(k)
		v := rv.GetParameterValueAsString(p)
		templateMap[k] = v
		os.Setenv(k, v)
	}
	b := make([]byte, 0, 8192)
	argresult := bytes.NewBuffer(b)
	//fmt.Printf("templateMap: %v\n", templateMap)
	err := argtemplate.Execute(argresult, templateMap)
	handleError(err)
	args := parseArgs(argresult.String())
	//fmt.Printf("args: %v\n", args)

	wf := *watchFile
	command := exec.Command(cmd, args...)
	//fmt.Printf("command: %v\nwf: %v\nwf == \"\": %v\n", command, wf, wf == "")
	if wf == "" {
		o, err := command.Output()
		obuf := bytes.NewReader(o)
		img, _, err := image.Decode(obuf)
		//handleError(err)
		if err != nil {
			obuf = bytes.NewReader(o)
			img, err = png.Decode(obuf)
		}
		//handleError(err)
		//fmt.Println("img is a type of:", reflect.TypeOf(img))
		if err != nil {
			return
		}
		m.Img = img
	} else {
		command.Run()
		f, err := os.Open(wf)
		if err != nil {
			return
		}
		//handleError(err)
		img, _, err := image.Decode(f)
		//handleError(err)
		f.Close()
		if err != nil {
			return
		}
		m.Img = img
	}
}

func parseArgs(argstring string) []string {
	//	fmt.Print(argstring)
	args := make([]string, 0, 10)
	b := make([]byte, 0, len(argstring))
	buf := bytes.NewBuffer(b)
	curquote := ""
	backslash := ""
	for _, s := range argstring {
		if s == ' ' && curquote == "" && backslash == "" {
			args = append(args, buf.String())
			//fmt.Printf("%v\n", args)
			buf.Reset()
		} else {
			buf.WriteRune(s)
		}
		if backslash == "" && curquote == "" {
			if s == '\'' || s == '"' || s == '`' {
				curquote = string(s)
			}
		} else if curquote != "" && curquote == string(s) && backslash == "" {
			curquote = ""
		}
		if s == '\\' && backslash == "" {
			backslash = "\\"
		} else {
			backslash = ""
		}
		//		fmt.Printf("s:%v curquote:%v backslash:%v\n", string(s), curquote, backslash)

	}
	if buf.Len() > 0 {
		args = append(args, buf.String())

	}
	return args

}

func createExtraFlags(extraflags string) []rv.RenderParameter {
	//	fmt.Printf("createExtraFlags: %v", extraflags)
	split := strings.Split(extraflags, ",")
	//	fmt.Print(split)
	if len(split) < 3 {
		flags := make([]rv.RenderParameter, 0, 0)
		return flags
	}
	flags := make([]rv.RenderParameter, 0, len(split)/3)
	for i := 0; i < (len(split) / 3); i++ {
		switch split[i*3+1] {
		case "int":
			iv, err := strconv.Atoi(split[i*3+2])
			handleError(err)
			flag := rv.NewIntRP(split[i*3], iv)
			flags = append(flags, flag)
		case "uint32":
			iv, err := strconv.ParseInt(split[i*3+2], 10, 32)
			handleError(err)
			flag := rv.NewUInt32RP(split[i*3], uint32(iv))
			flags = append(flags, flag)
		case "float64":
			fv, err := strconv.ParseFloat(split[i*3+2], 64)
			handleError(err)
			flag := rv.NewFloat64RP(split[i*3], fv)
			flags = append(flags, flag)
		case "complex128":
			cv, err := rv.ParseComplex(split[i*3+2])
			handleError(err)
			flag := rv.NewComplex128RP(split[i*3], cv)
			flags = append(flags, flag)
		default:
			flag := rv.NewStringRP(split[i*3], split[i*3+2])
			flags = append(flags, flag)

		}
	}
	//fmt.Print(flags)
	return flags

}

func handleError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
