package cmd

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

const templateFuncFolerName = "templates/func"

// check if the error is reported when the --lang arg is set incorrectly
func TestCreateLangErr(t *testing.T) {
	testFuncName := "test-create-func-c"
	rootCmd.SetArgs([]string{"create", testFuncName, "--lang", "c"})
	err := rootCmd.Execute()
	assert.Equal(t, fmt.Errorf("only supports cpp in this version"), err, "test create func failed: the expected error not be reported")
}

// check if the template downloaded from github is
// the same as the template already in `OpenRhino-CLI/templates`
func TestCreateFunc(t *testing.T) {
	// change test directory to ${workspaceFolder}
	// without this operation, contents downloaded from github will be saved in `cmd` directory
	os.Chdir("..")

	testFuncName := "test-create-func-cpp"
	rootCmd.SetArgs([]string{"create", testFuncName, "--lang", "cpp"})
	err := rootCmd.Execute()
	assert.Equal(t, nil, err, "test create func failed: %s", errorMessage(err))

	// check if the folder has been downloaded and unzipped successfully
	_, err = os.Stat(testFuncName)
	assert.Equal(t, nil, err, "test create func failed: %s", errorMessage(err))

	// read the 2 folder and check if they are exactly the same(filename, filecontent)
	checkDownloadFolerContent(t, testFuncName, templateFuncFolerName)

	// delete template folder
	err = os.RemoveAll(testFuncName)
	assert.Equal(t, nil, err, "test create func failed: %s", errorMessage(err))
}

func errorMessage(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func checkDownloadFolerContent(t *testing.T, downloadFolerName string, templateFolderName string) {
	// open and read 2 folders
	downloadFoler, err := os.Open(downloadFolerName)
	assert.Equal(t, nil, err, "test create func failed: %s", errorMessage(err))
	defer downloadFoler.Close()

	templateFolder, err := os.Open(templateFolderName)
	assert.Equal(t, nil, err, "test create func failed: %s", errorMessage(err))
	defer templateFolder.Close()

	downloadFolerInfo, err := downloadFoler.ReadDir(-1)
	assert.Equal(t, nil, err, "test create func failed: %s", errorMessage(err))
	templateFolderInfo, err := templateFolder.ReadDir(-1)
	assert.Equal(t, nil, err, "test create func failed: %s", errorMessage(err))

	// check if the number of entries in download folder and template folder are the same
	assert.Equal(t, len(templateFolderInfo), len(downloadFolerInfo), "number of entries in %s is not the same as "+
		"number of entries in %s", downloadFolerName, templateFolderName)

	// check if the entry names in the 2 folders are the same
	// if there are folders in the 2 folders, call this function recursively
	for i := 0; i < len(templateFolderInfo); i++ {
		// check entry name
		assert.Equal(t, downloadFolerInfo[i].Name(), templateFolderInfo[i].Name())

		downloadFileName := downloadFolerName + "/" + downloadFolerInfo[i].Name()
		templateFileName := templateFolderName + "/" + templateFolderInfo[i].Name()

		if downloadFolerInfo[i].IsDir() && templateFolderInfo[i].IsDir() {
			checkDownloadFolerContent(t, downloadFileName, templateFileName)
		} else {
			downloadFileContent, err := os.ReadFile(downloadFileName)
			assert.Equal(t, nil, err, "test create func failed: %s", errorMessage(err))
			templateFileContent, err := os.ReadFile(templateFileName)
			assert.Equal(t, nil, err, "test create func failed: %s", errorMessage(err))

			assert.Equal(t, string(templateFileContent), string(downloadFileContent),
				"test create func failed: file content of %s and %s are not the same",
				templateFileName, downloadFileName)
		}
	}
}
