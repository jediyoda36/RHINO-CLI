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
	"fmt"

	"github.com/spf13/cobra"
)

type DockerRunOptions struct {
	parallel int
	volume   string
}

func NewDockerRunCommand() *cobra.Command {
	dockerRunOpts := &DockerRunOptions{}
	dockerRunCmd := &cobra.Command{
		Use:   "docker-run [image]",
		Short: "Run an MPI program using Docker",
		Long:  "\nSubmit and run an MPI job using Docker",
		Example: `  rhino docker-run hello:v1.0
  rhino docker-run foo/matmul:v2.1 --np 4 -- arg1 arg2
  rhino docker-run bar/image:v3.0 -v /path/on/host:/path/in/container --np 8`,
		RunE: dockerRunOpts.dockerRun,
	}

	dockerRunCmd.Flags().StringVarP(&dockerRunOpts.volume, "volume", "v", "", "Bind mount a volume in the format <host-path>:<container-path>")
	dockerRunCmd.Flags().IntVar(&dockerRunOpts.parallel, "np", 1, "the number of MPI processes")

	return dockerRunCmd
}

func (r *DockerRunOptions) dockerRun(cmd *cobra.Command, args []string) error {
	// Check the arguments
	if len(args) == 0 {
		cmd.Help()
		return nil
	}
	if r.parallel < 1 {
		return fmt.Errorf("the number of MPI processes (--np) must be greater than 0")
	}

	// Create a DockerHelper instance
	helper, err := NewDockerHelper()
	if err != nil {
		return err
	}

	// Check if the image exists and pull it if necessary
	err = helper.checkAndPullImage(args[0])
	if err != nil {
		return err
	}

	// Create and start the container
	containerID, err := helper.createAndStartContainer(r, args)
	if err != nil {
		return err
	}

	// Get the container logs
	err = helper.getContainerLogs(containerID)
	if err != nil {
		return err
	}

	// Wait for the container to exit and retrieve the exit status
	return helper.waitForContainerExit(containerID)
}
