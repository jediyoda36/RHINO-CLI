package cmd

import (
	"os"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var RhinoJobGVR = schema.GroupVersionResource{Group: "openrhino.org", Version: "v1alpha1", Resource: "rhinojobs"}
var rootCmd = &cobra.Command{
	Use:   "rhino",
	Short: "\nRHINO-CLI - Manage your OpenRHINO functions and jobs",
}
var namespace string
var kubeconfig string

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}



