package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestNewDockerRunCommand(t *testing.T) {
	dockerRunCmd := NewDockerRunCommand()

	// Test if the returned object is of the correct type
	if _, ok := interface{}(dockerRunCmd).(*cobra.Command); !ok {
		t.Error("NewDockerRunCommand() did not return a *cobra.Command object")
	}
}

func TestDockerRun(t *testing.T) {
	// This test will require Docker to be installed and running on the test environment
	dockerRunCmd := NewDockerRunCommand()
	args := []string{"openrhino/rhino-test-sum:v0.1"}

	// TODO: Add more tests for different arguments
	err := dockerRunCmd.RunE(dockerRunCmd, args)
	if err != nil {
		t.Errorf("Docker run with 'openrhino/rhino-test-sum:v0.1' image failed: %s", err.Error())
	}
}
