package main

import (
	"github.com/andrew-d/go-termutil"
	"github.com/fatih/color"
	"os"
	"strconv"
	"text/template"
	"time"
)

const MergeRequestListTemplate string = `
{{ blue "#" }}{{ itoa .Iid | yellow }} {{ .Title | green | bold }}
{{ green .SourceBranch }} -> {{ red .TargetBranch }}

{{ .Description }}

`

const MergeRequestCheckoutListTemplate string = `{{ green .Title }}
`

const FeedTitleTemplate string = `
{{ .Title | bold  }}
`

const FeedTemplate string = `
{{ magenta "[" | bold  }}{{ .Updated | shortDate }}{{ magenta "]" | bold  }} {{ .Title }}
`

type formatFunc func(string, ...interface{}) string

// Map of colored template funcs: true, and non-colored: false
var templateFuncs map[bool]template.FuncMap

func init() {
	templateFuncs = make(map[bool]template.FuncMap)
	// Setup color functions for text/template
	colorFuncs := map[string]formatFunc{
		"green":   color.GreenString,
		"red":     color.RedString,
		"yellow":  color.YellowString,
		"white":   color.WhiteString,
		"cyan":    color.CyanString,
		"black":   color.BlackString,
		"blue":    color.BlueString,
		"magenta": color.MagentaString,
	}
	colorFuncMap := template.FuncMap{
		"bold": func(input string) string {
			return color.New(color.Bold).SprintFunc()(input)
		},
	}
	for c, fun := range colorFuncs {
		colorFuncMap[c] = func(finner formatFunc) func(string) string {
			return func(input string) string {
				return finner(input)
			}
		}(fun)
	}

	templateFuncs[true] = colorFuncMap

	monochromeFuncs := make(template.FuncMap)
	stringIdentity := func(input string) string {
		return input
	}
	for name, _ := range colorFuncMap {
		monochromeFuncs[name] = stringIdentity
	}
	templateFuncs[false] = monochromeFuncs

	// Shared functions
	for _, b := range []bool{true, false} {
		templateFuncs[b]["itoa"] = strconv.Itoa
		templateFuncs[b]["shortDate"] = func(t time.Time) string {
			return t.Format("02/01-06 15:04")
		}
	}
}

/// Determine from tty output, whether we should do colors
func doColors(output *os.File) bool {
	return termutil.Isatty(output.Fd())
}

/// Get new template for colored output
func newColorTemplate(name, format string) (*template.Template, error) {
	return newTemplate(name, format, true)
}

/// Get new template with monochrome version of color functions
func newMonochromeTemplate(name, format string) (*template.Template, error) {
	return newTemplate(name, format, false)
}

func newTemplate(name, format string, color bool) (*template.Template, error) {
	tmpl := template.New(name)
	tmpl.Funcs(templateFuncs[color])
	return tmpl.Parse(format)
}
