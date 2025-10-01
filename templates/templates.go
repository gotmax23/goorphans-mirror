package templates

import (
	"embed"
	"text/template"
)

//go:embed *.gotmpl
var templateFS embed.FS

var Templates = template.Must(template.ParseFS(templateFS, "*.gotmpl"))
