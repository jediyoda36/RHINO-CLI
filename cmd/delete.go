package cmd

import (
	"os"
	"fmt"
	"context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/homedir"
	"path/filepath"
	"github.com/spf13/cobra"
)

var rhinojobName string

var deleteCmd = &cobra.Command{
	Use:   "delete [name]",
	Short: "Delete rhino job and rhino function",
	Long: "\nDelete rhino job and rhino function",
	RunE: func(cmd *cobra.Command, args []string) error{
		if len(args) == 0 {
			return fmt.Errorf("[name] cannot be empty")
		}
		rhinojobName = args[0]
		var configPath string
		if len(kubeconfig) == 0 {
			if home := homedir.HomeDir(); home != "" {
				configPath = filepath.Join(home, ".kube", "config")
			} else {
				fmt.Println("Error: kubeconfig file not found, please use --config to specify the absolute path")
				os.Exit(0)
			}
		} else {
			configPath = kubeconfig
		}

		dynamicClient, currentNamespace, err := buildFromKubeconfig(configPath)
		if err != nil {
			return err
		}
		if namespace == "" {
			namespace = *currentNamespace
		}

		err = dynamicClient.Resource(RhinoJobGVR).Namespace(namespace).Delete(context.TODO(), rhinojobName, metav1.DeleteOptions{})
		if err != nil {
			fmt.Println("Error:", err.Error())
			os.Exit(0)
		}
		fmt.Println("RhinoJob " + rhinojobName + " deleted")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
	deleteCmd.Flags().StringVarP(&namespace, "namespace", "n", "", "namespace of the rhinojob")
	deleteCmd.Flags().StringVar(&kubeconfig, "kubeconfig", "", "kubernetes config path")
}
