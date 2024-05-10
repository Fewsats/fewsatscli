package store

import (
	"embed"
)

//go:embed migrations/*\.sql
var sqlSchemas embed.FS
