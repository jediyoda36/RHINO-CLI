package cmd

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var image string
var path string

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build MPI function/project",
	Long:  "\nBuild MPI function/project into a docker image",
	Example: `  rhino build ./hello.cpp --image foo/hello:v1.0
  rhino build /src/testbench -i bar/mpibench:v2.1`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 && len(image) == 0 {
			cmd.Help()
			os.Exit(0)
		} else if len(image) == 0 {
			return fmt.Errorf("please provide the image name")
		} else if len(args) == 0 {
			return fmt.Errorf("please provide the full path to your function or project folder")
		}
		path = args[0]
		if err := builder(image, path); err != nil {
			fmt.Println("Error:", err.Error())
			os.Exit(0)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)
	buildCmd.Flags().StringVarP(&image, "image", "i", "", "full image form: [registry]/[namespace]/[name]:[tag]")
}

func builder(image string, path string) error {
	f, err := os.Stat(path)
	if err != nil {
		return err
	}
	var execArgs []string
	funcName := getFuncName(image)

	// TODO: compile proj
	if f.IsDir() {
		out, err := execute("docker", []string{"images"})
		if err != nil {
			return err
		}
		fmt.Printf(out)
	} else {
		suffix := filepath.Ext(path)
		var compile string
		if suffix == ".c" {
			compile = "mpicc"
		} else if suffix == ".cpp" {
			compile = "mpic++"
		} else {
			return fmt.Errorf("only supports programs written in c or cpp")
		}

		execArgs = []string{
			"build", "-t", image,
			"--build-arg", "func_name=" + funcName,
			"--build-arg", "file=" + path,
			"--build-arg", "compile=" + compile,
			"-f", "./func.dockerfile", ".",
		}

		cmd := exec.Command("docker", execArgs...)
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return err
		}

		err = cmd.Start()
		if err != nil {
			return err
		}

		scanner := bufio.NewScanner(stdout)
		scanner.Split(bufio.ScanLines)
		for scanner.Scan() {
			cmdOutput := scanner.Text()
			fmt.Println(cmdOutput)
		}

		err = cmd.Wait()
		if err != nil {
			return err
		}
		// TODO: add image cleaner
	}
	return nil
}

func execute(commandName string, params []string) (string, error) {
	cmd := exec.Command(commandName, params...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Start()
	if err != nil {
		return "", err
	}
	err = cmd.Wait()
	return out.String(), err
}

func getFuncName(image string) string {
	nameTag := strings.Split(image, "/")
	funcName := strings.Split(nameTag[len(nameTag)-1], ":")[0]
	return funcName
}
