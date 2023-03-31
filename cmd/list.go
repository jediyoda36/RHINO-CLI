package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
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
		fmt.Println("Warning: no RhinoJobs found in the namespace")
	}

	var maxName float64 = 0
	var maxParallel float64 = 0
	var maxStatus float64 = 0
	for _, rj := range list.Items {
		// get max string length of rj.Name, rj.Spec.Parallelism, rj.Status.JobStatus
		maxName = math.Max(maxName, float64(len(string(rj.Name))))
		maxParallel = math.Max(maxParallel, float64(len(string(*rj.Spec.Parallelism))))
		maxStatus = math.Max(maxStatus, float64(len(string(rj.Status.JobStatus))))
	}
	nameFmt := fmt.Sprintf("%%-%ds", int(maxName)+20)
	parallelFmtD := fmt.Sprintf("%%-%dd", int(maxParallel)+15)
	parallelFmtS := fmt.Sprintf("%%-%ds", int(maxParallel)+15)
	statusFmt := fmt.Sprintf("%%-%ds", int(maxStatus)+5)

	fmt.Printf(nameFmt+parallelFmtS+statusFmt+"\n", "Name", "Parallelism", "Status")
	for _, rj := range list.Items {
		fmt.Printf(nameFmt+parallelFmtD+statusFmt+"\n", rj.Name, int(*rj.Spec.Parallelism), rj.Status.JobStatus)
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
