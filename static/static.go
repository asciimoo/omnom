// Package static provides embedded static assets for the Omnom web application.
//
// This package embeds CSS, JavaScript, images, icons, and other static files
// into the compiled binary using Go's embed directive. This allows Omnom to
// be distributed as a single executable with all assets included.
//
// The embedded files include:
//   - CSS stylesheets (including compiled SCSS)
//   - JavaScript files
//   - Web fonts
//   - Icons and images
//   - Documentation files
//
// User-generated content like snapshots and resources are NOT embedded and must
// be stored in a data directory specified in the configuration.
//
// The FS variable is an embed.FS that can be used with http.FileServer or other
// file-serving mechanisms.
package static

import "embed"

// Embed everything except the data directory which contains user-generated snapshot content

//go:embed css
//go:embed docs
//go:embed icons
//go:embed images
//go:embed js
//go:embed omnom.svg
//go:embed placeholder-image.png
//go:embed test
//go:embed webfonts

// FS contains all embedded static assets.
var FS embed.FS
