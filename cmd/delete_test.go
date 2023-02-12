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
	assert.Equal(t, nil, err, "test delete failed: %s", errorMessage(err))
	if strings.HasSuffix(cwd, "cmd") {
		os.Chdir("..")
	}

	// to test delete command
	// 1st: use `rhino create` to create a template folder
	// 2nd: use `rhino build` to build an image
	// 3rd: use `rhino run` to start a rhinojob
	testFuncName := "test-delete-func-cpp"
	rootCmd.SetArgs([]string{"create", testFuncName, "--lang", "cpp"})
	err = rootCmd.Execute()
	assert.Equal(t, nil, err, "preparatory work create failed: %s", errorMessage(err))

	os.Chdir(testFuncName)
	testFuncImageName := "test-delete-func-cpp:v1"
	rootCmd.SetArgs([]string{"build", "main.cpp", "--image", testFuncImageName})
	err = rootCmd.Execute()
	assert.Equal(t, nil, err, "preparatory work build failed: %s", errorMessage(err))

	// before exec `run` command, create a test namespace
	execute("kubectl", []string{"create", "namespace", testFuncRunNamespace})
	testRhinoJobName := "rhinojob-" + testFuncName
	rootCmd.SetArgs([]string{"run", testFuncImageName, "--namespace", testFuncRunNamespace})
	err = rootCmd.Execute()
	assert.Equal(t, nil, err, "preparatory work run failed: %s", errorMessage(err))

	// test delete
	rootCmd.SetArgs([]string{"delete", testRhinoJobName})
	err = rootCmd.Execute()
	assert.Equal(t, nil, err, "test delete failed: %s", errorMessage(err))

	// check if the rhinojob created just now is deleted successfully
	actualCmdOutput, err := execute("kubectl", []string{"get", "rhinojob", "-n", testFuncRunNamespace})
	assert.Equal(t, nil, err, "test delete failed: %s", errorMessage(err))

	expetedCmdOutput := "No resources found in test-func-cpp-ns namespace.\n"
	assert.Equal(t, expetedCmdOutput, actualCmdOutput, "test delete failed:\n"+
		"expected kubectl output: %s\nactual kubectl output: %s\n",
		expetedCmdOutput, actualCmdOutput)

	// delete test namespace and rhinojob created just now
	execute("kubectl", []string{"delete", "rhinojob", testRhinoJobName, "-n", testFuncRunNamespace})
	execute("kubectl", []string{"delete", "namespace", testFuncRunNamespace})

	// delete the image built just now
	execute("docker", []string{"rmi", testFuncImageName})

	// remove template folder
	os.Chdir("..")
	os.RemoveAll(testFuncName)
}
