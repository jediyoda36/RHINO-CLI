package cmd

import (
	"os"
	"fmt"
	"context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"path/filepath"
	"github.com/spf13/cobra"
)

var name string
var namespaced string

var deleteCmd = &cobra.Command{
	Use:   "delete [name]",
	Short: "Delete rhino job and rhino function",
	Long: "\nDelete rhino job and rhino function",
	RunE: func(cmd *cobra.Command, args []string) error{
		if len(args) == 0 {
			return fmt.Errorf("[name] cannot be empty")
		}
		name = args[0]
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
		config, err := clientcmd.BuildConfigFromFlags("", configPath)
		if err != nil {
			return err
		}
	
		dynamicClient, err := dynamic.NewForConfig(config)
		if err != nil {
			return err
		}
	
		err = dynamicClient.Resource(RhinoJobGVR).Namespace(namespaced).Delete(context.TODO(), name, metav1.DeleteOptions{})
		if err != nil {
			fmt.Println("Error:", err.Error())
			os.Exit(0)
		}
		fmt.Println("RhinoJob " + name + " deleted")
		return nil	
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
	deleteCmd.Flags().StringVarP(&namespaced, "namespace", "n", "default", "namespace of the rhinojob")
	deleteCmd.Flags().StringVar(&kubeconfig, "kubeconfig", "", "kubernetes config path")
}
