package cmd

import (
	"bytes"
	"io"
	"os"
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

func TestDockerRunWithNoArgs(t *testing.T) {
	// This test will require Docker to be installed and running on the test environment
	dockerRunCmd := NewDockerRunCommand()
	args := []string{"openrhino/rhino-test-sum:v0.1"}

	// Run the command and check for errors
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

func TestDockerRunWithArgs(t *testing.T) {
	dockerRunCmd := NewDockerRunCommand()

	args := []string{"openrhino/integration:test-v0.2.0", "1", "1", "1"} //Set the arguments
	dockerRunCmd.Flags().Set("np", "2")                                  //Set the Flags.
	//For this test image, the number of MPI processes
	//should be equal or greater than 2

	// Run the command and check for errors
	err := dockerRunCmd.RunE(dockerRunCmd, args)
	if err != nil {
		t.Errorf("Rhino docker-run with 'openrhino/integration:test-v0.2.0' image failed: %s", err.Error())
	}
}

func TestDockerRunWithInvalidFlag(t *testing.T) {
	dockerRunCmd := NewDockerRunCommand()
	args := []string{"openrhino/integration:test-v0.2.0", "1", "1", "1"}
	dockerRunCmd.Flags().Set("np", "1") //Set the Flags.
	//For this test image, the number of MPI processes
	//should be equal or greater than 2, so this will fail

	// Redirect os.Stderr to a buffer
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	err := dockerRunCmd.RunE(dockerRunCmd, args)

	// Close the writer and restore os.Stderr
	w.Close()
	os.Stderr = oldStderr

	// Read the output from the buffer
	var buf bytes.Buffer
	io.Copy(&buf, r)

	// Error is expected because the flag is invalid
	if err == nil {
		t.Error("Rhino docker-run with invalid flag did not return an error")
	} else {
		output := buf.String()
		expectedOutputSubstring := "Function needs at least two processes"
		if !strings.Contains(output, expectedOutputSubstring) {
			t.Errorf("Rhino docker-run output does not contain the expected error message.\nGot:\n%s\nExpected to contain:\n%s", output, expectedOutputSubstring)
			expectedErrorSubstring := "container exited with non-zero status"
			if !strings.Contains(err.Error(), expectedErrorSubstring) {
				t.Errorf("Rhino docker-run with invalid args returned unexpected error: %s", err.Error())
			}
		}
	}
}

func TestDockerRunWithInvalidArgs(t *testing.T) {
	dockerRunCmd := NewDockerRunCommand()
	args := []string{"openrhino/integration:test-v0.2.0", "1"} //Set the arguments to be invalid
	//For this test image, the MPI function requires 3 parameters, so this will fail
	dockerRunCmd.Flags().Set("np", "2") //Set the Flags.
	//For this test image, the number of MPI processes
	//should be equal or greater than 2
	err := dockerRunCmd.RunE(dockerRunCmd, args)

	// Error is expected because the args are invalid
	if err == nil {
		t.Error("Rhino docker-run with invalid args did not return an error")
	} else {
		expectedErrorSubstring := "container exited with non-zero status"
		if !strings.Contains(err.Error(), expectedErrorSubstring) {
			t.Errorf("Rhino docker-run with invalid args returned unexpected error: %s", err.Error())
		}
	}
}