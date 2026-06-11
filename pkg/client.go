package pkg

import (
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
)

type RESTClientGetter interface {
	// ToRESTConfig returns restconfig
	ToRESTConfig() (*rest.Config, error)
	// ToDiscoveryClient returns discovery client
	ToDiscoveryClient() (discovery.CachedDiscoveryInterface, error)
	// ToRESTMapper returns a restmapper
	ToRESTMapper() (meta.RESTMapper, error)
	// ToRawKubeConfigLoader return kubeconfig loader as-is
	ToRawKubeConfigLoader() clientcmd.ClientConfig
}

type Client struct {
	config      *rest.Config
	clientConfig clientcmd.ClientConfig
	discoveryClient discovery.CachedDiscoveryInterface
	restMapper    meta.RESTMapper
}

func NewClient() (*Client, error) {
	c := &Client{}

	// Initialize client config
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	c.clientConfig = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, &clientcmd.ConfigOverrides{})

	return c, nil
}

func (c *Client) ToRESTConfig() (*rest.Config, error) {
	if c.config != nil {
		return c.config, nil
	}

	config, err := c.clientConfig.ClientConfig()
	if err != nil {
		return nil, err
	}
	c.config = config
	return config, nil
}

func (c *Client) ToDiscoveryClient() (discovery.CachedDiscoveryInterface, error) {
	if c.discoveryClient != nil {
		return c.discoveryClient, nil
	}

	config, err := c.ToRESTConfig()
	if err != nil {
		return nil, err
	}

	discoveryClient, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return nil, err
	}

	c.discoveryClient = memory.NewMemCacheClient(discoveryClient)
	return c.discoveryClient, nil
}

func (c *Client) ToRESTMapper() (meta.RESTMapper, error) {
	if c.restMapper != nil {
		return c.restMapper, nil
	}

	discoveryClient, err := c.ToDiscoveryClient()
	if err != nil {
		return nil, err
	}

	groupResources, err := restmapper.GetAPIGroupResources(discoveryClient)
	if err != nil {
		return nil, err
	}

	c.restMapper = restmapper.NewDiscoveryRESTMapper(groupResources)
	return c.restMapper, nil
}

func (c *Client) ToRawKubeConfigLoader() clientcmd.ClientConfig {
	return c.clientConfig
}
