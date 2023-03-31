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
		return fmt.Errorf(err.Error())
	}
	fmt.Println("RhinoJob " + d.rhinojobName + " deleted")
	return nil
}
