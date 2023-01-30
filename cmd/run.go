package cmd

import (
	"fmt"
	"os"
	"strconv"
	"context"
	"encoding/json"
	"path/filepath"
	"github.com/spf13/cobra"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"k8s.io/apimachinery/pkg/runtime/schema"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var imageName string
var execArgs []string
var parallel int
var execTime int
var dataPath string
var dataServer string
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

var runCmd = &cobra.Command{
	Use:   "run [image]",
	Short: "Submit and run rhino job",
	Long: "\nSubmit and run rhino job",
	Example: `  rhino run hello:v1.0 --namespace user_space
  rhino run foo/matmul:v2.1 --np 4 -- arg1 arg2 
  rhino run mpi/testbench -n 32 -t 800 --server 10.0.0.7 --dir /mnt -- --in=/data/file --out=/data/out`,
	RunE: func(cmd *cobra.Command, args []string) error{
		if len(args) == 0 {
			cmd.Help()
			os.Exit(0)
		}
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
		run, err := runRhinoJob(dynamicClient, namespace, args)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(0)
		}
		fmt.Println(run)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().StringVar(&dataServer, "server", "", "NFS server ip")
	runCmd.Flags().StringVar(&dataPath, "dir", "", "shared directory in NFS server")
	runCmd.MarkFlagsRequiredTogether("server", "dir")
	runCmd.Flags().IntVarP(&parallel, "np", "n", 1, "mpi processes")
	runCmd.Flags().IntVarP(&execTime, "ttl", "t", 600, "estimated execution time(s)")
	runCmd.Flags().StringVar(&namespace, "namespace", "default", "namespace of the rhinojob")
	runCmd.Flags().StringVar(&config, "config", "", "kubernetes config path")
}

func printYAML(args []string) (yamlFile string) {
	funcName := getFuncName(args[0])
	yamlFile = `apiVersion: openrhino.org/v1alpha1
kind: RhinoJob
metadata:
  labels:
    app.kubernetes.io/name: rhinojob 
    app.kubernetes.io/instance: rhinojob-` 
	yamlFile += funcName + `
    app.kubernetes.io/part-of: rhino-operator
    app.kubernetes.io/managed-by: kustomize
    app.kubernetes.io/created-by: rhino-operator
  name: rhinojob-`
  	yamlFile += funcName +`
spec:
  image: "`
  	yamlFile += args[0] + `"
  ttl: `
  	yamlFile += strconv.Itoa(execTime) + `
  parallelism: `
  	yamlFile += strconv.Itoa(parallel) + ` 
  appExec: "./` 
  	yamlFile += funcName
	if len(args) > 1 {
		yamlFile += `"
  appArgs: [`
		for i := 1; i < len(args); i++ {
			yamlFile += `"` + args[i] + `", `
		}
		yamlFile += `]`	
	} 
	if len(dataServer) != 0 {
		yamlFile += `
  dataServer: "` + dataServer + `"`
  		yamlFile += `
  dataPath: "` + dataPath + `"`
	}
	return yamlFile
}

func runRhinoJob(client dynamic.Interface, namespace string, args []string) (*RhinoJob, error) {
	decoder := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	obj := &unstructured.Unstructured{}
	if _, _, err := decoder.Decode([]byte(printYAML(args)), nil, obj); err != nil {
		return nil, err
	}

	create, err := client.Resource(gvr).Namespace(namespace).Create(context.TODO(), obj, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}
	data, err := create.MarshalJSON()
	if err != nil {
		return nil, err
	}
	var rj RhinoJob
	if err := json.Unmarshal(data, &rj); err != nil {
		return nil, err
	}
	return &rj, nil
}