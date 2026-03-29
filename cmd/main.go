package main

import (
	"fmt"
	"os"

	"github.com/nevtum/nimbletask/pkg/cli"
)

// main executes the root command
func main() {
	cmd := cli.NewRootCmd()
	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
