// Package templates contains html templates and embeds those in variable
package templates

import (
	"embed"
	"html/template"
)

//go:embed *.html
var templateFS embed.FS

// Templates contains all html templates inside folder package
var Templates = template.Must(template.ParseFS(templateFS, "*.html"))
