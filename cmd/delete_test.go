package cmd

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDeleteSingleJob(t *testing.T) {
	// change work directory to ${workspaceFolder}
	cwd, err := os.Getwd()
	assert.Equal(t, nil, err, "test delete failed: %s", errorMessage(err))
	if strings.HasSuffix(cwd, "cmd") {
		os.Chdir("..")
	}
	// use `rhino build` to build template
	os.Chdir("templates/func")
	testFuncName := "test-delete-func-cpp"
	testFuncImageName := "test-delete-func-cpp:v1"
	rootCmd.SetArgs([]string{"build", "--image", testFuncImageName})
	err = rootCmd.Execute()
	assert.Equal(t, nil, err, "preparatory work build failed: %s", errorMessage(err))

	// test run command
	execute("kubectl", []string{"create", "namespace", testFuncRunNamespace})
	rootCmd.SetArgs([]string{"run", testFuncImageName, "--namespace", testFuncRunNamespace})
	err = rootCmd.Execute()
	assert.Equal(t, nil, err, "preparatory work run failed: %s", errorMessage(err))

	// test delete
	fmt.Println("Wait 10s and check job status")
	time.Sleep(10 * time.Second)
	testRhinoJobName := "rhinojob-" + testFuncName
	rootCmd.SetArgs([]string{"delete", testRhinoJobName, "--namespace", testFuncRunNamespace})
	err = rootCmd.Execute()
	assert.Equal(t, nil, err, "test delete failed: %s", errorMessage(err))

	// check if the rhinojob created just now is deleted successfully
	actualCmdOutput, err := execute("kubectl", []string{"get", "rhinojob", "-n", testFuncRunNamespace})
	assert.Equal(t, nil, err, "test delete failed: %s", errorMessage(err))

	expetedCmdOutput := "No resources found in rhino-test namespace.\n"
	assert.Equal(t, expetedCmdOutput, actualCmdOutput, "test delete failed:\n"+
		"expected kubectl output: %s\nactual kubectl output: %s\n",
		expetedCmdOutput, actualCmdOutput)

	// delete test namespace and rhinojob created just now
	execute("kubectl", []string{"delete", "namespace", testFuncRunNamespace, "--force", "--grace-period=0"})
	
	// delete the image built just now
	execute("docker", []string{"rmi", testFuncImageName})
	execute("sh", []string{"-c", "docker rmi -f $(docker images | grep none | grep second | awk '{print $3}')"})
}
