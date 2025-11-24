// Package docs provides embedded documentation files for Omnom.
//
// This package embeds Markdown documentation into the compiled binary, making
// documentation accessible without requiring separate files. The documentation
// includes:
//   - API reference
//   - User guides
//   - Installation instructions
//   - Configuration examples
//
// Documentation files can be served via HTTP or accessed programmatically for
// in-app help systems.
package docs

import "embed"

// Embed everything except the data directory which contains user-generated snapshot content

// FS contains embedded documentation files.
//
//go:embed *.md
var FS embed.FS
