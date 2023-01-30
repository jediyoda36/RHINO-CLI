package cmd

import (
	"os"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

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

var RhinoJobGVR = schema.GroupVersionResource{Group: "openrhino.org", Version: "v1alpha1", Resource: "rhinojobs"}
var rootCmd = &cobra.Command{
	Use:   "rhino",
	Short: "\nRHINO-CLI - Manage your OpenRHINO functions and jobs",
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
}


