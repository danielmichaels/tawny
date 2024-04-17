package assets

import (
	"embed"
)

// AppName is the application name used across the repository.
const AppName = "tawny"

//go:embed "assets/static" "gen"
var EmbeddedFiles embed.FS
