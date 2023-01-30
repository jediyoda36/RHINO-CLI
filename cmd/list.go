package cmd

import (
	"os"
	"fmt"
	"context"
	"encoding/json"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"path/filepath"
	"github.com/spf13/cobra"
)

var namespace string
var config string
var gvr = schema.GroupVersionResource{Group: "openrhino.org", Version: "v1alpha1", Resource: "rhinojobs"}

type RhinoJobSpec struct {
	Image string `json:"image"`
	Parallelism *int32 `json:"parallelism,omitempty"`
	TTL        *int32   `json:"ttl,omitempty"`
	AppExec    string   `json:"appExec"`
	AppArgs    []string `json:"appArgs,omitempty"`
	DataServer string   `json:"dataServer,omitempty"`
	DataPath   string   `json:"dataPath,omitempty"`
}

type RhinoJobStatus struct {
	JobStatus JobStatus `json:"jobStatus"`
}
type JobStatus string

const (
	Pending   JobStatus = "Pending"
	Running   JobStatus = "Running"
	Failed    JobStatus = "Failed"
	Completed JobStatus = "Completed"
)

type RhinoJob struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RhinoJobSpec   `json:"spec,omitempty"`
	Status RhinoJobStatus `json:"status,omitempty"`
}

type RhinoJobList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RhinoJob `json:"items"`
}


var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all rhino jobs",
	Long: "\nList all rhino jobs",
	Example: `  rhino list
  rhino list --namespace user_func`,
	RunE: func(cmd *cobra.Command, args []string) error{
		var kubeconfig string

		if home := homedir.HomeDir(); home != "" {
			kubeconfig = filepath.Join(home, ".kube", "config")
		} else {
			if len(config) == 0 {
				fmt.Println("Error: kubeconfig file not found, please use --config to specify the absolute path")
				os.Exit(0)
			}
			kubeconfig = config
		}

		config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return err
		}
	
		dynamicClient, err := dynamic.NewForConfig(config)
		if err != nil {
			return err
		}
	
		list, err := listRhinoJob(dynamicClient, namespace)
		if err != nil {
			return err
		}
		if len(list.Items) == 0 {
			fmt.Println("RhinoJob not found in namespace:", namespace)
			os.Exit(0)
		}
		fmt.Printf("%-20s\t%-15s\t%-5s\n", "Name", "Parallelism", "Status")
		for _, rj := range list.Items {
			fmt.Printf("%-20v\t%-15v\t%-5v\n", rj.Name, *rj.Spec.Parallelism, rj.Status.JobStatus)
		}
		return nil	
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().StringVarP(&namespace, "namespace", "n", "default", "namespace of the rhinojob")
	listCmd.Flags().StringVar(&config, "config", "", "kubernetes config path")
}

func listRhinoJob(client dynamic.Interface, namespace string) (*RhinoJobList, error) {
	list, err := client.Resource(gvr).Namespace(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
			return nil, err
	}
	data, err := list.MarshalJSON()
	if err != nil {
			return nil, err
	}
	var rjList RhinoJobList
	if err := json.Unmarshal(data, &rjList); err != nil {
			return nil, err
	}
	return &rjList, nil
}