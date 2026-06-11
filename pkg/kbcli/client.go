package kbcli

import (
	"github.com/apecloud/kubeblocks/pkg/client/clientset/versioned"
	"helm.sh/helm/v3/pkg/cli"
	"k8s.io/client-go/kubernetes"
)

func getClient() (*versioned.Clientset, error) {
	cli := cli.New()
	restConfig, err := cli.RESTClientGetter().ToRESTConfig()
	if err != nil {
		return nil, err
	}
	return versioned.NewForConfig(restConfig)
}
func getDefaultClient() (*kubernetes.Clientset, error) {
	cli := cli.New()
	restConfig, err := cli.RESTClientGetter().ToRESTConfig()
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(restConfig)
}
