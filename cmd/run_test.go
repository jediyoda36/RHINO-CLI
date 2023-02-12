package cmd

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

const testFuncRunNamespace = "test-func-cpp-ns"

func TestRunSingleJob(t *testing.T) {
	// change work directory to ${workspaceFolder}
	cwd, err := os.Getwd()
	assert.Equal(t, nil, err, "test run failed: %s", errorMessage(err))
	if strings.HasSuffix(cwd, "cmd") {
		os.Chdir("..")
	}

	// to test run command
	// first: use `rhino create` to create a template folder
	// second: use `rhino build` to build an image
	testFuncName := "test-run-func-cpp"
	rootCmd.SetArgs([]string{"create", testFuncName, "--lang", "cpp"})
	err = rootCmd.Execute()
	assert.Equal(t, nil, err, "preparatory work create failed: %s", errorMessage(err))

	os.Chdir(testFuncName)
	testFuncImageName := "test-run-func-cpp:v1"
	rootCmd.SetArgs([]string{"build", "main.cpp", "--image", testFuncImageName})
	err = rootCmd.Execute()
	assert.Equal(t, nil, err, "preparatory work build failed: %s", errorMessage(err))

	// before test `run` command, create a test namespace
	execute("kubectl", []string{"create", "namespace", testFuncRunNamespace})

	// test run command
	rootCmd.SetArgs([]string{"run", testFuncImageName, "--namespace", testFuncRunNamespace})
	err = rootCmd.Execute()
	assert.Equal(t, nil, err, "test run failed: %s", errorMessage(err))

	// use `kubectl get rhinojob` to check whether rhinojob has been created
	cmdOutput, err := execute("kubectl", []string{"get", "rhinojob", "--namespace", testFuncRunNamespace})
	assert.Equal(t, nil, err, "test run failed: %s", errorMessage(err))

	cmdOutputLines := strings.Split(cmdOutput, "\n")
	testRhinoJobName := "rhinojob-" + testFuncName
	var foundTestRhinoJob bool
	for _, line := range cmdOutputLines {
		if strings.HasPrefix(line, testRhinoJobName) {
			foundTestRhinoJob = true
			break
		}
	}
	assert.Equal(t, true, foundTestRhinoJob, "test run failed: rhinojob not found")

	// delete test namespace and rhinojob created just now
	execute("kubectl", []string{"delete", "rhinojob", testRhinoJobName, "-n", testFuncRunNamespace})
	execute("kubectl", []string{"delete", "namespace", testFuncRunNamespace})

	// delete the image built just now
	execute("docker", []string{"rmi", testFuncImageName})

	// remove template folder
	os.Chdir("..")
	os.RemoveAll(testFuncName)
}
