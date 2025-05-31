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
var FS embed.FS
