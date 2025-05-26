package main

import (
	"embed"

	"github.com/asciimoo/omnom/cmd"
)

//go:embed "config.yml_sample"
var fs embed.FS

func main() {
	cmd.Execute(fs)
}
