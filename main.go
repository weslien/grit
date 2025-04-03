package main

import (
	"github.com/weslien/grit/cmd"
)

var Version = "0.1.0" // This will be set during build

func main() {
	cmd.Execute(Version)
}
