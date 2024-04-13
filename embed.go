package assets

import (
	"embed"
)

//go:embed "assets/static" "gen/http"
var EmbeddedFiles embed.FS
