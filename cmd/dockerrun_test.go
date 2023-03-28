package cmd

import (
	"strings"
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
		t.Errorf("Rhino docker-run with 'openrhino/rhino-test-sum:v0.1' image failed: %s", err.Error())
	}
}

func TestDockerRunWithNonExistentImage(t *testing.T) {
	dockerRunCmd := NewDockerRunCommand()
	args := []string{"openrhino/do-not-exist"}

	err := dockerRunCmd.RunE(dockerRunCmd, args)
	// Error is expected because the image does not exist
	if err == nil {
		t.Error("Rhino docker-run with nonexistent image did not return an error")
	} else {
		expectedErrorSubstring := "not exist"
		if !strings.Contains(err.Error(), expectedErrorSubstring) {
			t.Errorf("Rhino docker-run with nonexistent image returned unexpected error: %s", err.Error())
		}
	}
}

func TestDockerRunWithInvalidArg(t *testing.T) {
	dockerRunCmd := NewDockerRunCommand()
	args := []string{"openrhino/integration", "invalid-arg"}

	err := dockerRunCmd.RunE(dockerRunCmd, args)
	// Error is expected because the image does not exist
	if err == nil {
		t.Error("Rhino docker-run with invalid args did not return an error")
	} else {
		expectedErrorSubstring := "container exited with non-zero status"
		if !strings.Contains(err.Error(), expectedErrorSubstring) {
			t.Errorf("Rhino docker-run with invalid args returned unexpected error: %s", err.Error())
		}
	}
}
