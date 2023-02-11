package cmd

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildSingleFileCpp(t *testing.T) {
	// change work directory to ${workspaceFolder}
	cwd, err := os.Getwd()
	assert.Equal(t, nil, err, "test build failed: %s", errorMessage(err))
	if strings.HasSuffix(cwd, "cmd") {
		os.Chdir("..")
	}

	// to test build, first create a template folder
	// this is a unit test for `build` command, so no further detailed test on `create`
	testFuncName := "test-build-func-cpp"
	rootCmd.SetArgs([]string{"create", testFuncName, "--lang", "cpp"})
	err = rootCmd.Execute()
	assert.Equal(t, nil, err, "preparatory work create failed: %s", errorMessage(err))

	// change work directory to template folder
	os.Chdir(testFuncName)

	testFuncImageName := "test-build-func-cpp:v1"
	rootCmd.SetArgs([]string{"build", "main.cpp", "--image", testFuncImageName})
	err = rootCmd.Execute()
	assert.Equal(t, nil, err, "test build failed: %s", errorMessage(err))

	// use `docker image` to check whether the image has been built
	cmdOutput, err := execute("docker", []string{"images"})
	assert.Equal(t, err, nil, "test build failed when using `docker images`: %s", errorMessage(err))
	cmdOutputLines := strings.Split(cmdOutput, "\n")

	var foundBuiltImage bool
	for _, line := range cmdOutputLines {
		if strings.HasPrefix(line, testFuncName) {
			foundBuiltImage = true
		}
	}
	assert.Equal(t, true, foundBuiltImage, "test build failed: could not find the test image built using `docker images`")

	// remove test image
	if foundBuiltImage {
		execute("docker", []string{"rmi", testFuncImageName})
	}

	// remove template folder
	os.Chdir("..")
	os.RemoveAll(testFuncName)
}
