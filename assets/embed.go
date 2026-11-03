// Package assets embeds static application assets.
package assets

import _ "embed"

// IconPNG holds the embedded application notification icon.
//
//go:embed icon.png
var IconPNG []byte
