package cmd

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeleteSingleJob(t *testing.T) {
	// change work directory to ${workspaceFolder}
	cwd, err := os.Getwd()
	assert.Equal(t, nil, err, "test list failed: %s", errorMessage(err))
	if strings.HasSuffix(cwd, "cmd") {
		os.Chdir("..")
	}
	// use `rhino build` to build integration sample
	os.Chdir("samples/integration")
	testFuncName := "test-delete-func-cpp"
	testFuncImageName := "test-delete-func-cpp:v1"
	rootCmd.SetArgs([]string{"build", "--image", testFuncImageName})
	err = rootCmd.Execute()
	assert.Equal(t, nil, err, "preparatory work build failed: %s", errorMessage(err))

	// test run command
	rootCmd.SetArgs([]string{"run", testFuncImageName, "--np", "2", "--", "1", "10", "1"})
	err = rootCmd.Execute()
	assert.Equal(t, nil, err, "preparatory work run failed: %s", errorMessage(err))

	// test delete
	testRhinoJobName := "rhinojob-" + testFuncName
	rootCmd.SetArgs([]string{"delete", testRhinoJobName})
	err = rootCmd.Execute()
	assert.Equal(t, nil, err, "test delete failed: %s", errorMessage(err))

	// check if the rhinojob created just now is deleted successfully
	actualCmdOutput, err := execute("kubectl", []string{"get", "rhinojob"})
	assert.Equal(t, nil, err, "test delete failed: %s", errorMessage(err))

	expetedCmdOutput := "No resources found in"
	assert.Equal(t, true, strings.Contains(actualCmdOutput, expetedCmdOutput), "test delete failed:\n"+
		"expected kubectl output start with: %s\nactual kubectl output: %s\n",
		expetedCmdOutput, actualCmdOutput)

	// delete rhinojob created just now
	execute("kubectl", []string{"delete", "rhinojob", testRhinoJobName})

	// delete the image built just now
	execute("docker", []string{"rmi", testFuncImageName})
	execute("sh", []string{"-c", "docker rmi -f $(docker images | grep none | grep second | awk '{print $3}')"})
}
