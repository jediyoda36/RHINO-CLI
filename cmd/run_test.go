package cmd

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRunSingleJob(t *testing.T) {
	// change work directory to ${workspaceFolder}
	cwd, err := os.Getwd()
	assert.Equal(t, nil, err, "test run failed: %s", errorMessage(err))
	if strings.HasSuffix(cwd, "cmd") {
		os.Chdir("..")
	}

	// use `rhino build` to build integration sample
	os.Chdir("samples/integration")
	testFuncName := "test-run-func-cpp"
	testFuncImageName := "test-run-func-cpp:v1"
	rootCmd.SetArgs([]string{"build", "--image", testFuncImageName})
	err = rootCmd.Execute()
	assert.Equal(t, nil, err, "preparatory work build failed: %s", errorMessage(err))

	// test run command
	rootCmd.SetArgs([]string{"run", testFuncImageName, "--np", "2", "--", "1", "10", "1"})
	err = rootCmd.Execute()
	assert.Equal(t, nil, err, "test run failed: %s", errorMessage(err))

	// use `kubectl get rhinojob` to check whether rhinojob has been created
	fmt.Println("Waiting 10s...")
	time.Sleep(10 * time.Second)
	cmdOutput, err := execute("kubectl", []string{"get", "rhinojob"})
	assert.Equal(t, nil, err, "test run failed: %s", errorMessage(err))
	assert.Equal(t, true, strings.Contains(cmdOutput, "Running"), "rhinojob failed to start")
	
	// delete rhinojob created just now
	testRhinoJobName := "rhinojob-" + testFuncName
	execute("kubectl", []string{"delete", "rhinojob", testRhinoJobName})

	// delete the image built just now
	execute("docker", []string{"rmi", testFuncImageName})
	execute("sh", []string{"-c", "docker rmi -f $(docker images | grep none | grep second | awk '{print $3}')"})
}
