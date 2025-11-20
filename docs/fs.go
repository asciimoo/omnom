package docs

import "embed"

// Embed everything except the data directory which contains user-generated snapshot content

//go:embed *.md
var FS embed.FS
