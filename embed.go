package assets

import (
	"embed"
)

//go:embed "assets/static"  "assets/view" "gen/http"
var EmbeddedFiles embed.FS
