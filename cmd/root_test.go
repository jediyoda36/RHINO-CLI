package cmd_test

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/OpenRHINO/RHINO-CLI/cmd"
)

func TestNewRootCommand(t *testing.T) {
	rootCmd := cmd.NewRootCommand()

	// Test if rootCmd has the correct Use, Short, and Long values
	assert.Equal(t, "rhino", rootCmd.Use)
	assert.Equal(t, "\nRHINO-CLI - Manage your OpenRHINO functions and jobs", rootCmd.Short)

	// Test if rootCmd has the correct subcommands
	expectedSubcommands := []string{"create", "build", "delete", "run", "list", "docker-run"}
	actualSubcommands := getSubcommandNames(rootCmd)

	assert.Equal(t, len(expectedSubcommands), len(actualSubcommands), "Number of subcommands should be equal")

	for _, expected := range expectedSubcommands {
		assert.Contains(t, actualSubcommands, expected, "Expected subcommand %s is not present", expected)
	}
}

func getSubcommandNames(cmd *cobra.Command) []string {
	subcommandNames := []string{}
	for _, subCmd := range cmd.Commands() {
		subcommandNames = append(subcommandNames, subCmd.Name())
	}
	return subcommandNames
}
