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
	"github.com/spf13/cobra"
)


type DockerRunOptions struct {
	parallel   int
	volume    string
	funcName   string
}


func NewDockerRunCommand() *cobra.Command {
	dockerRunOpts := &DockerRunOptions{}
	dockerRunCmd := &cobra.Command{
		Use:   "docker-run [image]",
		Short: "Submit and run a RHINO job using Docker",
		Long:  "\nSubmit an MPI function/project and run it as a RHINO job using Docker",
		Example: `  rhino dockerRun hello:v1.0
  rhino dockerRun foo/matmul:v2.1 --np 4 -- arg1 arg2 `,
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
	// 处理 -v 参数
	hostPath, containerPath := "", ""
	if r.volume != "" {
		volumeParts := strings.SplitN(r.volume, ":", 2)
		if len(volumeParts) == 2 {
			hostPath, containerPath = volumeParts[0], volumeParts[1]
		} else {
			return fmt.Errorf("invalid volume format, should be <host-path>:<container-path>")
		}
	}

	// 设置上下文和Docker客户端
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}

	// Configure the container
	entrypoint := []string{"mpirun", "-np", strconv.Itoa(r.parallel), "/app/mpi-func"}
	containerConfig := &container.Config{
		Image: args[0],
		Entrypoint: entrypoint,
		Cmd:   args[1:],
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

	// Attach container output to the command line program's output
	attachOptions := types.ContainerAttachOptions{
		Stream: true,
		Stdout: true,
		Stderr: true,
	}
	attachResp, err := cli.ContainerAttach(ctx, resp.ID, attachOptions)
	if err != nil {
		return err
	}
	defer attachResp.Close()

	// Copy container output to stdout
	_, err = io.Copy(os.Stdout, attachResp.Reader)
	if err != nil && err != io.EOF {
		return err
	}

	return nil
}
