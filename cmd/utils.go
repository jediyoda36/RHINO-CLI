package cmd
import (
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
)

func buildFromKubeconfig(configPath string) (dynamicClient *dynamic.DynamicClient, currentNamespace *string, err error) {
	// We use 2 kinds of config here.
	// The dynamicClient need to be constructed with rest.Config.
	// On the other hand, we need to use api.Config or ClientConfig to
	// read the context info and current namespace from the kubeconfig file.
	// The rest.Config does not include the namespace.
	config, err := clientcmd.BuildConfigFromFlags("", configPath)
	if err != nil {
		return nil, nil, err
	}
	dynamicClient, err = dynamic.NewForConfig(config)
	if err != nil {
		return nil, nil, err
	}

	cmdapiConfig, err := clientcmd.LoadFromFile(configPath)
	if err != nil {
		return nil, nil, err
	}
	context, exist := cmdapiConfig.Contexts[cmdapiConfig.CurrentContext]
	if exist {
		if context.Namespace == "" { 
			//If namespace is not defined in kubeconfig, use "default"
			context.Namespace = "default"
		}
		currentNamespace = &context.Namespace
	} else {
		return nil, nil, err
	}

	return dynamicClient, currentNamespace, nil
}