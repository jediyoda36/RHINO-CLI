package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

var image string
var file string

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build MPI function/project",
	Long:  "\nBuild MPI function/project into a docker image",
	Example: `  rhino build --image foo/hello:v1.0
  rhino build -f ./src/config/Makefile -i bar/mpibench:v2.1 -- make -j all arch=Linux`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(image) == 0 {
			return fmt.Errorf("please provide the image name")
		} else if len(args) > 0 && args[0] != "make" {
			return fmt.Errorf("build command must start with 'make'")
		}
		// check image name
		validName := regexp.MustCompile("^[a-z0-9]([-a-z0-9]*[a-z0-9])?$")
		matchString := validName.MatchString(getFuncName(image))
		if !matchString {
			return fmt.Errorf("image name can only contain a~z, 0~9 and -")
		}

		if err := builder(args, image, file); err != nil {
			fmt.Println("Error:", err.Error())
			os.Exit(0)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)
	buildCmd.Flags().StringVarP(&image, "image", "i", "", "full image form: [registry]/[namespace]/[name]:[tag]")
	buildCmd.Flags().StringVarP(&file, "file", "f", "", "relative path of the makefile")
}

func builder(args []string, image string, file string) error {
	var execArgs []string
	var execCommand string
	var buildCommand []string = []string{"make"}
	var makefilePath string
	var funcName string = "mpi-func"

	// check Makefile
	if len(file) == 0 {
		makefilePath = "./src/Makefile"
	} else {
		makefilePath = file
	}
	_, err := os.Stat(makefilePath)
	if err != nil {
		return err
	}
	fmt.Println("Makefile path:", makefilePath)

	// add build args
	if len(args) > 0 {
		buildCommand = args
	}
	fmt.Println("Build command:", buildCommand)

	// check build tools
	buildFiles := []string{"Dockerfile", "ldd.sh"}
	for _, file := range buildFiles {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			return fmt.Errorf("build template not found. Please use 'rhino create' first")
		}
	}
	fmt.Println("Build tools found. Start building...")

	execCommand = "docker"
	execArgs = []string{
		"build", "-t", image,
		"--rm",
		"--build-arg", "func_name=" + funcName,
		"--build-arg", "file=" + makefilePath,
		"--build-arg", "make_args=" + strings.Join(buildCommand[1:], " "),
		".",
	}

	cmd := exec.Command(execCommand, execArgs...)
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	err = cmd.Start()
	if err != nil {
		return err
	}

	go printPipeOutput(stdoutPipe)
	go printPipeOutput(stderrPipe)

	err = cmd.Wait()
	if err != nil {
		return err
	}
	return nil
}

func printPipeOutput(pipe io.ReadCloser) {
	scanner := bufio.NewScanner(pipe)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		output := scanner.Text()
		fmt.Println(output)
	}
}
