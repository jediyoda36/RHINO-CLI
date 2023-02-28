package cmd

import (
	"bufio"
	"bytes"
	"fmt"
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

	if len(args) > 0 {
		buildCommand = args
	}
	fmt.Println("Build command:", buildCommand)		

	_, err = os.Stat("Dockerfile") 
	if err != nil {
		return fmt.Errorf("build template not found. Please use 'rhino create' first")
	}

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