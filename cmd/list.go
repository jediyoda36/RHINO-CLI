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
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	rhinojob "github.com/OpenRHINO/RHINO-Operator/api/v1alpha1"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/util/homedir"
)

type ListOptions struct {
	kubeconfig string
	namespace  string
}

func NewListCommand() *cobra.Command {
	listOpts := &ListOptions{}
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List RHINO jobs",
		Long:  "\nList all the RHINO jobs in your current namespace or the namespace specified",
		Example: `  rhino list
  rhino list --namespace user_func`,
		RunE: listOpts.list,
	}

	listCmd.Flags().StringVarP(&listOpts.namespace, "namespace", "n", "", "the namespace to list RHINO jobs")
	listCmd.Flags().StringVar(&listOpts.kubeconfig, "kubeconfig", "", "the path of the kubeconfig file")

	return listCmd
}

func (l *ListOptions) list(cmd *cobra.Command, args []string) error {
	// Check the arguments
	if len(args) != 0 {
		cmd.Help()
		return nil
	}

	// Get the kubeconfig file
	if len(l.kubeconfig) == 0 {
		if home := homedir.HomeDir(); home != "" {
			l.kubeconfig = filepath.Join(home, ".kube", "config")
		} else {
			return fmt.Errorf("kubeconfig file not found, please use --config to specify the absolute path")
		}
	}

	// Build the dynamic client
	dynamicClient, currentNamespace, err := buildFromKubeconfig(l.kubeconfig)
	if err != nil {
		return err
	}
	if l.namespace == "" {
		l.namespace = *currentNamespace
	}
	list, err := l.listRhinoJob(dynamicClient)
	if err != nil {
		return err
	}
	if len(list.Items) == 0 {
		fmt.Println("Warning: no RhinoJobs found in the namespace")
	} else {
		w := tabwriter.NewWriter(os.Stdout, 2, 0, 2, ' ', 0)
		fmt.Fprintln(w, "Name\tParallelism\tStatus")

		for _, rj := range list.Items {
			fmt.Fprintf(w, "%s\t%d\t%s\n", rj.Name, *rj.Spec.Parallelism, rj.Status.JobStatus)
		}

		// 刷新输出，确保所有内容写入 stdout
		if err := w.Flush(); err != nil {
			return err
		}
	}

	return nil
}

func (l *ListOptions) listRhinoJob(client dynamic.Interface) (*rhinojob.RhinoJobList, error) {
	list, err := client.Resource(RhinoJobGVR).Namespace(l.namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	data, err := list.MarshalJSON()
	if err != nil {
		return nil, err
	}
	var rjList rhinojob.RhinoJobList
	if err := json.Unmarshal(data, &rjList); err != nil {
		return nil, err
	}
	return &rjList, nil
}
