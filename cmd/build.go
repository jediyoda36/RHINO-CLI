/*
 * Copyright 2023 RHINO Team
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
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

type BuildOptions struct {
	image string
	file  string
}

func NewBuildCommand() *cobra.Command {
	buildOpts := &BuildOptions{}

	buildCmd := &cobra.Command{
		Use:   "build",
		Short: "Build MPI function/project",
		Long:  "\nBuild MPI function/project into a docker image",
		Example: `  rhino build --image foo/hello:v1.0
  rhino build -f ./src/config/Makefile -i bar/mpibench:v2.1 -- make -j all arch=Linux`,
		Args: buildOpts.validateArgs,
		RunE: buildOpts.runBuild,
	}

	buildCmd.Flags().StringVarP(&buildOpts.image, "image", "i", "", "full image form: [registry]/[namespace]/[name]:[tag]")
	buildCmd.Flags().StringVarP(&buildOpts.file, "file", "f", "", "relative path of the makefile")

	return buildCmd
}

func (b *BuildOptions) validateArgs(buildCmd *cobra.Command, args []string) error {
	if len(b.image) == 0 {
		return fmt.Errorf("please provide the image name")
	} else if len(b.image) > 63 {
		return fmt.Errorf("the image name cannot exceed 63 characters")
	} else if len(args) > 0 && args[0] != "make" {
		return fmt.Errorf("build command must start with 'make'")
	}

	validName := regexp.MustCompile("^[a-z0-9]([-a-z0-9]*[a-z0-9])?$")
	matchString := validName.MatchString(getFuncName(b.image))
	if !matchString {
		return fmt.Errorf("image name can only contain a~z, 0~9 and -")
	}
	return nil
}

func (b *BuildOptions) runBuild(buildCmd *cobra.Command, args []string) error {
	var execArgs []string
	var execCommand string
	var buildCommand []string = []string{"make"}
	var makefilePath string
	var funcName string = "mpi-func" //TODO: Since the funcName is hard coded and used in different files, remove this line and use an external const to avoid inconsistency.

	// check Makefile
	if len(b.file) == 0 {
		makefilePath = "./src/Makefile"
	} else {
		makefilePath = b.file
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
		"build", "-t", b.image,
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
