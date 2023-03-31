package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"

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
		return fmt.Errorf("no RhinoJobs found in the namespace")
	}
	fmt.Printf("%-20s\t%-15s\t%-5s\n", "Name", "Parallelism", "Status")
	for _, rj := range list.Items {
		fmt.Printf("%-20v\t%-15v\t%-5v\n", rj.Name, *rj.Spec.Parallelism, rj.Status.JobStatus)
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
