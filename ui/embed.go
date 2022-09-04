//go:build embedui

package ui

import (
	"embed"
	"io/fs"

	"github.com/labstack/echo/v4"
)

//go:embed dist/*
var embedded embed.FS

var EmbeddedUI fs.FS

func init() {
	EmbeddedUI = echo.MustSubFS(embedded, "dist/")
}
