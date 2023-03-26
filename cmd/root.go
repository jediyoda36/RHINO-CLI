package cmd

import (
	"github.com/spf13/cobra"
)

func NewRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "rhino",
		Short: "\nRHINO-CLI - Manage your OpenRHINO functions and jobs",
	}

	rootCmd.AddCommand(NewCreateCommand())
	rootCmd.AddCommand(NewBuildCommand())
	rootCmd.AddCommand(NewDeleteCommand())
	rootCmd.AddCommand(NewRunCommand())
	rootCmd.AddCommand(NewListCommand())
	rootCmd.AddCommand(NewDockerRunCommand())

	return rootCmd
}