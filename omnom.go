// Omnom is a self-hosted bookmark manager and feed aggregator with ActivityPub support.
//
// Omnom allows you to save and organize web bookmarks with:
//   - Full-page snapshots with diff tracking
//   - Hierarchical collections for organization
//   - Powerful tagging and search
//   - Public and private bookmarks
//   - Browser extension support
//
// It also functions as a feed reader supporting:
//   - RSS/Atom feeds
//   - ActivityPub federation (Mastodon, Pleroma, etc.)
//   - Per-user read/unread tracking
//   - Content aggregation
//
// Key features:
//   - Passwordless authentication via email
//   - OAuth support (GitHub, Google, OIDC)
//   - REST API for integrations
//   - Content preservation with snapshots
//   - Multi-user support
//   - Internationalization
//
// Usage:
//
//	# Start the server
//	omnom listen
//
//	# Create a user
//	omnom create-user username email@example.com
//
//	# Add a bookmark from CLI
//	omnom create-bookmark username "Title" https://example.com
//
//	# Update feeds
//	omnom update-feeds
//
// For more information, see: https://github.com/asciimoo/omnom
package main

import (
	"embed"

	"github.com/asciimoo/omnom/cmd"
)

//go:embed "config.yml_sample"
var fs embed.FS

func main() {
	cmd.Execute(fs)
}
