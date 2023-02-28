package cmd

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListSingleJob(t *testing.T) {
	// change work directory to ${workspaceFolder}
	cwd, err := os.Getwd()
	assert.Equal(t, nil, err, "test list failed: %s", errorMessage(err))
	if strings.HasSuffix(cwd, "cmd") {
		os.Chdir("..")
	}
	// use `rhino build` to build integration sample
	os.Chdir("samples/integration")
	testFuncName := "test-list-func-cpp"
	testFuncImageName := "test-list-func-cpp:v1"
	rootCmd.SetArgs([]string{"build", "--image", testFuncImageName})
	err = rootCmd.Execute()
	assert.Equal(t, nil, err, "preparatory work build failed: %s", errorMessage(err))

	// test run command
	rootCmd.SetArgs([]string{"run", testFuncImageName, "--np", "2", "--", "1", "10", "1"})
	err = rootCmd.Execute()
	assert.Equal(t, nil, err, "preparatory work run failed: %s", errorMessage(err))

	// test list command
	// when calling `rootCmd.Execute` to execute `list` command
	// rhinoJob info will be sent directly to os.Stdout and thus cannot be collected
	// in order to collected that info, we replace os.Stdout with pipes in unix, and then
	// switch back after checking finishes
	rescueStdout := os.Stdout
	r, w, err := os.Pipe()
	assert.Equal(t, nil, err, "test list failed: %s", errorMessage(err))

	os.Stdout = w
	rootCmd.SetArgs([]string{"list"})
	err = rootCmd.Execute()
	assert.Equal(t, nil, err, "test list failed: %s", errorMessage(err))

	// switch back
	w.Close()
	os.Stdout = rescueStdout

	// copy the output of `rhino list` from read end of the pipe to a string builder
	buf := new(strings.Builder)
	io.Copy(buf, r)
	r.Close()

	cmdOutput := buf.String()
	cmdOutputLines := strings.Split(cmdOutput, "\n")

	var foundRhinoJob bool
	testRhinoJobName := "rhinojob-" + testFuncName
	for _, line := range cmdOutputLines {
		if strings.HasPrefix(line, testRhinoJobName) {
			foundRhinoJob = true
			break
		}
	}
	assert.Equal(t, true, foundRhinoJob, "test list failed: list output does not contain rhinojob created in this test")

	// delete rhinojob created just now
	execute("kubectl", []string{"delete", "rhinojob", testRhinoJobName})

	// delete the image built just now
	execute("docker", []string{"rmi", testFuncImageName})
	execute("sh", []string{"-c", "docker rmi -f $(docker images | grep none | grep second | awk '{print $3}')"})
}
