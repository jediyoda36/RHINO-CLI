package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/spf13/cobra"
)

type DockerRunOptions struct {
	parallel int
	volume   string
	funcName string
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
		os.Exit(0)
	}
	r.funcName = getFuncName(args[0])
	if r.parallel < 1 {
		return fmt.Errorf("the number of MPI processes (--np) must be greater than 0")
	}
	// Handle -v option
	hostPath, containerPath := "", ""
	if r.volume != "" {
		volumeParts := strings.SplitN(r.volume, ":", 2)
		if len(volumeParts) == 2 {
			hostPath, containerPath = volumeParts[0], volumeParts[1]
		} else {
			return fmt.Errorf("invalid volume format, should be <host-path>:<container-path>")
		}
	}

	// Set up the context and Docker client
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}

	// Check if the image exists
	_, _, err = cli.ImageInspectWithRaw(ctx, args[0])
	if err != nil {
		if client.IsErrNotFound(err) {
			fmt.Printf("Image %s not found, pulling from Docker Hub...\n", args[0])
			// Pull the image
			out, err := cli.ImagePull(ctx, args[0], types.ImagePullOptions{})
			if err != nil {
				return err
			}
			defer out.Close()
			// Copy the output to stdout
			_, err = io.Copy(os.Stdout, out)
			if err != nil && err != io.EOF {
				return err
			}
		} else {
			return err
		}
	}

	// Configure the container
	entrypoint := []string{"mpirun", "-np", strconv.Itoa(r.parallel), "/app/mpi-func"}
	containerConfig := &container.Config{
		Image:      args[0],
		Entrypoint: entrypoint,
		Cmd:        args[1:],
		Env: []string{
			"OMPI_MCA_btl_base_warn_component_unused=0", // Suppress OpenMPI warning, if OpenMPI is used
		},
	}
	hostConfig := &container.HostConfig{}
	// Bind mount a volume if "-v" is specified
	if hostPath != "" && containerPath != "" {
		bind := []string{hostPath + ":" + containerPath}
		hostConfig.Binds = bind
	}

	// Create and start the container
	resp, err := cli.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, "")
	if err != nil {
		return err
	}
	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return err
	}

	// Get the container logs
	logOptions := types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
	}
	logReader, err := cli.ContainerLogs(ctx, resp.ID, logOptions)
	if err != nil {
		return err
	}
	defer logReader.Close()

	// Use a demultiplexer to split stdout and stderr, and copy the container logs to the program output
	_, err = stdcopy.StdCopy(os.Stdout, os.Stderr, logReader)
	if err != nil && err != io.EOF {
		return fmt.Errorf("error copying container logs: %v", err)
	}

	// Wait for the container to exit and retrieve the exit status
	waitCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	var waitResp container.WaitResponse
	var waitErr error

	select {
	case waitResp = <-waitCh:
		if waitResp.StatusCode != 0 {
			return fmt.Errorf("container exited with non-zero status: %d", waitResp.StatusCode)
		}
	case waitErr = <-errCh:
		return waitErr
	}

	return nil
}
