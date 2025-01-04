package apps

import (
	"embed"
)

//go:embed all:*/*.go all:*/**/*.go
var SourceFiles embed.FS

//go:embed all:*/**/*.html all:*/**/*.sql
var ResourceFiles embed.FS
