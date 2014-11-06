package main

import (
	"github.com/codegangsta/cli"
	"github.com/fatih/color"
	"strconv"
	"text/template"
)

const MergeRequestListTemplate string = `
{{ blue "#" }}{{ itoa .Iid | yellow }} {{ .Title | green | bold }}
{{ green .SourceBranch }} -> {{ red .TargetBranch }}

{{ .Description }}

`

const MergeRequestCheckoutListTemplate string = `{{ green .Title }}
`

type formatFunc func(string, ...interface{}) string

var TermTemplateFuncMap map[string]interface{}

// Get new template with defaults
func newTemplate(c *cli.Context, name, format string) (*template.Template, error) {

	if nil == TermTemplateFuncMap {
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
		m := map[string]interface{}{
			"bold": func(input string) string {
				return color.New(color.Bold).SprintFunc()(input)
			},
			"itoa": strconv.Itoa,
		}
		for c, fun := range colorFuncs {
			m[c] = func(finner formatFunc) func(string) string {
				return func(input string) string {
					return finner(input)
				}
			}(fun)
		}

		TermTemplateFuncMap = m
	}
	tmpl := template.New(name)
	tmpl.Funcs(TermTemplateFuncMap)
	return tmpl.Parse(format)
}
