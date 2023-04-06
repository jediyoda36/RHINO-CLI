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
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/homedir"
)

type DeleteOptions struct {
	rhinojobName string
	kubeconfig   string
	namespace    string
}

func NewDeleteCommand() *cobra.Command {
	deleteOpts := &DeleteOptions{}
	deleteCmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a RHINO job by name",
		Long:  "\nDelete a RHINO job by name",
		Args:  deleteOpts.argsCheck,
		RunE:  deleteOpts.runDelete,
	}
	deleteCmd.Flags().StringVarP(&deleteOpts.namespace, "namespace", "n", "", "namespace of the RHINO job")
	deleteCmd.Flags().StringVar(&deleteOpts.kubeconfig, "kubeconfig", "", "path to the kubeconfig file")

	return deleteCmd
}

func (d *DeleteOptions) argsCheck(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("[name] cannot be empty")
	}
	d.rhinojobName = args[0]
	if len(d.kubeconfig) == 0 {
		if home := homedir.HomeDir(); home != "" {
			d.kubeconfig = filepath.Join(home, ".kube", "config")
		} else {
			return fmt.Errorf("kubeconfig file not found, please use --config to specify the absolute path")
		}
	}

	return nil
}

func (d *DeleteOptions) runDelete(cmd *cobra.Command, args []string) error {
	dynamicClient, currentNamespace, err := buildFromKubeconfig(d.kubeconfig)
	if err != nil {
		return err
	}
	if d.namespace == "" {
		d.namespace = *currentNamespace
	}

	err = dynamicClient.Resource(RhinoJobGVR).Namespace(d.namespace).Delete(context.TODO(), d.rhinojobName, metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	fmt.Println("RhinoJob " + d.rhinojobName + " deleted")
	return nil
}
