// Package ui serves the embedded production dashboard frontend.
//
// The dist directory is generated build input: it is staged from
// frontend/dist by `make ui-stage` and is never committed. Compilation
// fails clearly when the staged assets are absent, so the binary can
// never ship with an empty UI.
package ui

import "embed"

//go:embed all:dist
var distFS embed.FS
