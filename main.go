package main

import (
	"os"

	"github.com/kwojcicki/ldapbeat/cmd"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
