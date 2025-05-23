package templates

import "embed"

//go:embed *.tpl */*.tpl
//go:embed *.xml
var FS embed.FS
