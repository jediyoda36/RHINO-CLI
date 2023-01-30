package cmd

import (
	"os"
	"fmt"
	"context"
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
var name string
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


var deleteCmd = &cobra.Command{
	Use:   "delete [name]",
	Short: "Delete rhino job and rhino function",
	Long: "\nDelete rhino job and rhino function",
	RunE: func(cmd *cobra.Command, args []string) error{
		if len(args) == 0 {
			return fmt.Errorf("[name] cannot be empty")
		}
		name = args[0]
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
	
		err = dynamicClient.Resource(gvr).Namespace(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(0)
		}
		fmt.Println("RhinoJob " + name + " deleted")
		return nil	
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
	deleteCmd.Flags().StringVarP(&namespace, "namespace", "n", "default", "namespace of the rhinojob")
	deleteCmd.Flags().StringVar(&config, "config", "", "kubernetes config path")
}
