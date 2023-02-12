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

	// to test list command
	// 1st: use `rhino create` to create a template folder
	// 2nd: use `rhino build` to build an image
	// 3rd: use `rhino run` to start a rhinojob
	testFuncName := "test-list-func-cpp"
	rootCmd.SetArgs([]string{"create", testFuncName, "--lang", "cpp"})
	err = rootCmd.Execute()
	assert.Equal(t, nil, err, "preparatory work create failed: %s", errorMessage(err))

	os.Chdir(testFuncName)
	testFuncImageName := "test-list-func-cpp:v1"
	rootCmd.SetArgs([]string{"build", "main.cpp", "--image", testFuncImageName})
	err = rootCmd.Execute()
	assert.Equal(t, nil, err, "preparatory work build failed: %s", errorMessage(err))

	// before exec `run` command, create a test namespace
	execute("kubectl", []string{"create", "namespace", testFuncRunNamespace})
	testRhinoJobName := "rhinojob-" + testFuncName
	rootCmd.SetArgs([]string{"run", testFuncImageName, "--namespace", testFuncRunNamespace})
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
	rootCmd.SetArgs([]string{"list", "--namespace", testFuncRunNamespace})
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
	for _, line := range cmdOutputLines {
		if strings.HasPrefix(line, testRhinoJobName) {
			foundRhinoJob = true
			break
		}
	}
	assert.Equal(t, true, foundRhinoJob, "test list failed: list output does not contain rhinojob created in this test")

	// delete test namespace and rhinojob created just now
	execute("kubectl", []string{"delete", "rhinojob", testRhinoJobName, "-n", testFuncRunNamespace})
	execute("kubectl", []string{"delete", "namespace", testFuncRunNamespace})

	// delete the image built just now
	execute("docker", []string{"rmi", testFuncImageName})

	// remove template folder
	os.Chdir("..")
	os.RemoveAll(testFuncName)
}
