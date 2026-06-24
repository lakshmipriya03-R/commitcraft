package main

import (
	"os"

	"github.com/lakshmipriya03-R/commitcraft/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
