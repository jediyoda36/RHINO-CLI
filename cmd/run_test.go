package cmd

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const testFuncRunNamespace = "rhino-test"

func TestRunSingleJob(t *testing.T) {
	// change work directory to ${workspaceFolder}
	cwd, err := os.Getwd()
	assert.Equal(t, nil, err, "test run failed: %s", errorMessage(err))
	if strings.HasSuffix(cwd, "cmd") {
		os.Chdir("..")
	}
	rootCmd := NewRootCommand()
	// use `rhino build` to build template
	os.Chdir("templates/func")
	testFuncImageName := "test-run-func-cpp:v1"
	rootCmd.SetArgs([]string{"build", "--image", testFuncImageName})
	err = rootCmd.Execute()
	assert.Equal(t, nil, err, "preparatory work build failed: %s", errorMessage(err))

	// test run command
	execShellCmd("kubectl", []string{"create", "namespace", testFuncRunNamespace})
	rootCmd.SetArgs([]string{"run", testFuncImageName, "--namespace", testFuncRunNamespace})
	err = rootCmd.Execute()
	assert.Equal(t, nil, err, "test run failed: %s", errorMessage(err))

	// use `kubectl get rhinojob` to check whether rhinojob has been created
	fmt.Println("Wait 10s and check job status")
	time.Sleep(10 * time.Second)
	cmdOutput, err := execShellCmd("kubectl", []string{"get", "rhinojob", "--namespace", testFuncRunNamespace})
	assert.Equal(t, nil, err, "test run failed: %s", errorMessage(err))
	assert.Equal(t, true, strings.Contains(cmdOutput, "Completed"), "rhinojob failed to start")

	// delete rhinojob created just now
	execShellCmd("kubectl", []string{"delete", "namespace", testFuncRunNamespace, "--force", "--grace-period=0"})

	// delete the image built just now
	execShellCmd("docker", []string{"rmi", testFuncImageName})
	execShellCmd("sh", []string{"-c", "docker rmi -f $(docker images | grep none | grep second | awk '{print $3}')"})
}
