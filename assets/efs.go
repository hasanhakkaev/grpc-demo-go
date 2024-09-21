package assets

import "embed"

//go:embed "migrations" "configuration.yaml"
var EmbeddedFiles embed.FS
