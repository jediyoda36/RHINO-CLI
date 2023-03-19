package main

import (
	"github.com/OpenRHINO/RHINO-CLI/cmd"
)

func main() {
	rootCmd := cmd.NewRootCommand()
	rootCmd.Execute()
}
