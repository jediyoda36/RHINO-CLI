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
	"context"
	"strings"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
)

var RhinoJobGVR = schema.GroupVersionResource{Group: "openrhino.org", Version: "v1alpha1", Resource: "rhinojobs"}

func buildFromKubeconfig(configPath string) (dynamicClient *dynamic.DynamicClient, currentNamespace *string, err error) {
	// We use 2 kinds of config here.
	// The dynamicClient need to be constructed with rest.Config.
	// On the other hand, we need to use api.Config or ClientConfig to
	// read the context info and current namespace from the kubeconfig file.
	// The rest.Config does not include the namespace.
	config, err := clientcmd.BuildConfigFromFlags("", configPath)
	if err != nil {
		return nil, nil, err
	}
	dynamicClient, err = dynamic.NewForConfig(config)
	if err != nil {
		return nil, nil, err
	}

	cmdapiConfig, err := clientcmd.LoadFromFile(configPath)
	if err != nil {
		return nil, nil, err
	}
	context, exist := cmdapiConfig.Contexts[cmdapiConfig.CurrentContext]
	if exist {
		if context.Namespace == "" { 
			//If namespace is not defined in kubeconfig, use "default"
			context.Namespace = "default"
		}
		currentNamespace = &context.Namespace
	} else {
		return nil, nil, err
	}

	return dynamicClient, currentNamespace, nil
}

func getFuncName(image string) string {
	nameTag := strings.Split(image, "/")
	funcName := strings.Split(nameTag[len(nameTag)-1], ":")[0]
	return funcName
}


// DockerHelper is a helper struct for Docker operations
type DockerHelper struct {
	ctx context.Context
	cli *client.Client
}

// The factory function creates a new DockerHelper instance
func NewDockerHelper() (*DockerHelper, error) {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}

	return &DockerHelper{
		ctx: ctx,
		cli: cli,
	}, nil
}

func (dh *DockerHelper) checkAndPullImage(image string) error {
	_, _, err := dh.cli.ImageInspectWithRaw(dh.ctx, image)
	if err != nil {
		if client.IsErrNotFound(err) {
			fmt.Printf("Image %s not found, pulling from Docker Hub...\n", image)
			// Pull the image
			out, err := dh.cli.ImagePull(dh.ctx, image, types.ImagePullOptions{})
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
	return nil
}

func (dh *DockerHelper) createAndStartContainer(r *DockerRunOptions, args []string) (string, error) {
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
	hostPath, containerPath := "", ""
	if r.volume != "" {
		volumeParts := strings.SplitN(r.volume, ":", 2)
		if len(volumeParts) == 2 {
			hostPath, containerPath = volumeParts[0], volumeParts[1]
		} else {
			return "", fmt.Errorf("invalid volume format, should be <host-path>:<container-path>")
		}
	}
	if hostPath != "" && containerPath != "" {
		bind := []string{hostPath + ":" + containerPath}
		hostConfig.Binds = bind
	}

	// Create and start the container
	resp, err := dh.cli.ContainerCreate(dh.ctx, containerConfig, hostConfig, nil, nil, "")
	if err != nil {
		return "", err
	}
	if err := dh.cli.ContainerStart(dh.ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return "", err
	}

	return resp.ID, nil
}

func (dh *DockerHelper) getContainerLogs(containerID string) error {
	logOptions := types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
	}
	logReader, err := dh.cli.ContainerLogs(dh.ctx, containerID, logOptions)
	if err != nil {
		return err
	}
	defer logReader.Close()

	// Use a demultiplexer to split stdout and stderr, and copy the container logs to the program output
	_, err = stdcopy.StdCopy(os.Stdout, os.Stderr, logReader)
	if err != nil && err != io.EOF {
		return fmt.Errorf("error copying container logs: %v", err)
	}

	return nil
}

func (dh *DockerHelper) waitForContainerExit(containerID string) error {
	waitCh, errCh := dh.cli.ContainerWait(dh.ctx, containerID, container.WaitConditionNotRunning)
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
