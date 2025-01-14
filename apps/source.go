package apps

import (
	"embed"
)

//go:embed all:*/*.go all:*/**/*.go
var SourceFiles embed.FS

//go:embed all:*/templates/*.html all:*/templates/**/*.html all:*/migrations/*.sql
var ResourceFiles embed.FS
