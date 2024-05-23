package messages

import (
	_ "embed"
	"io"
	"text/template"
)

var (
	//go:embed messages.gotmpl
	rawMessageTpl string

	messageTpl *template.Template

	funcMap = template.FuncMap{
		"sub": func(a, b int) int {
			return a - b
		},
		"add": func(a, b int) int {
			return a + b
		},
	}
)

func init() {
	messageTpl = template.Must(template.New("messages").Funcs(funcMap).Parse(rawMessageTpl))
}

func Write(msg *Messages, pkg string, w io.Writer) error {
	return messageTpl.Execute(w, map[string]any{
		"Package":       pkg,
		"Messages":      msg,
		"VarTypeInt":    VarTypeInt,
		"VarTypeString": VarTypeString,
		"VarTypeFloat":  VarTypeFloat,
	})
}
