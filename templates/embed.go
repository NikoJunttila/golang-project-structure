// internal/templates/embed.go
package templates

import (
	"embed"
	"html/template"
)

//go:embed *.html
var templateFS embed.FS

var Templates = template.Must(template.ParseFS(templateFS, "*.html"))
