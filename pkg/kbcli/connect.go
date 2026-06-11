package kbcli

import (
	"context"
	"fmt"
	"sort"
	"strconv"

	kbappsv1 "github.com/apecloud/kubeblocks/apis/apps/v1"
	"github.com/apecloud/kubeblocks/pkg/client/clientset/versioned"
	"github.com/apecloud/kubeblocks/pkg/constant"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
)

type ConnectWrapper struct {
	namespace     string
	clusterName   string
	componentName string
	instanceName  string
	clientType    string

	kubeClient kubernetes.Interface
	kbClient   versioned.Interface
}

type ConnectInfoJSON struct {
	Namespace          string                `json:"namespace"`
	ClusterName        string                `json:"clusterName"`
	ComponentName      string                `json:"componentName,omitempty"`
	InstanceName       string                `json:"instanceName,omitempty"`
	ClientType         string                `json:"clientType,omitempty"`
	ServiceKind        string                `json:"serviceKind,omitempty"`
	PortForwardCommand string                `json:"portForwardCommand,omitempty"`
	ClientExample      string                `json:"clientExample,omitempty"`
	Endpoints          []ConnectEndpointJSON `json:"endpoints"`
	Accounts           []ConnectAccountJSON  `json:"accounts"`
}

type ConnectEndpointJSON struct {
	ComponentName string            `json:"componentName,omitempty"`
	ShardingName  string            `json:"shardingName,omitempty"`
	ServiceName   string            `json:"serviceName"`
	Type          string            `json:"type"`
	Internal      string            `json:"internal"`
	External      string            `json:"external,omitempty"`
	Ports         []ConnectPortJSON `json:"ports"`
}

type ConnectPortJSON struct {
	Name     string `json:"name,omitempty"`
	Port     int32  `json:"port"`
	NodePort int32  `json:"nodePort,omitempty"`
}

type ConnectAccountJSON struct {
	ComponentName string `json:"componentName,omitempty"`
	ShardingName  string `json:"shardingName,omitempty"`
	SecretName    string `json:"secretName"`
	Username      string `json:"username"`
	Password      string `json:"password"`
	PasswordKey   string `json:"passwordKey"`
}

type componentRef struct {
	componentName string
	shardingName  string
	componentDef  string
}

func NewConnectWrapper(namespace, clusterName, componentName, instanceName, clientType string) (*ConnectWrapper, error) {
	if namespace == "" {
		namespace = "default"
	}
	if clusterName == "" && instanceName == "" {
		return nil, fmt.Errorf("either cluster or instance must be specified")
	}
	kubeClient, err := getDefaultClient()
	if err != nil {
		return nil, err
	}
	kbClient, err := getClient()
	if err != nil {
		return nil, err
	}
	return &ConnectWrapper{
		namespace:     namespace,
		clusterName:   clusterName,
		componentName: componentName,
		instanceName:  instanceName,
		clientType:    clientType,
		kubeClient:    kubeClient,
		kbClient:      kbClient,
	}, nil
}

func (w *ConnectWrapper) ToJSON() (*ConnectInfoJSON, error) {
	ctx := context.TODO()

	if w.instanceName != "" {
		pod, err := w.kubeClient.CoreV1().Pods(w.namespace).Get(ctx, w.instanceName, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		if w.clusterName == "" {
			w.clusterName = pod.Labels[constant.AppInstanceLabelKey]
		}
		if w.componentName == "" {
			w.componentName = pod.Labels[constant.KBAppComponentLabelKey]
		}
	}

	clusterObj, err := w.kbClient.AppsV1().Clusters(w.namespace).Get(ctx, w.clusterName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	services, err := w.listClusterServices(ctx)
	if err != nil {
		return nil, err
	}
	targetServices := w.selectServices(clusterObj, services)
	if len(targetServices) == 0 {
		return nil, fmt.Errorf("cannot find any available services")
	}

	componentRefs, err := w.resolveComponentRefs(ctx, clusterObj, targetServices)
	if err != nil {
		return nil, err
	}

	resp := &ConnectInfoJSON{
		Namespace:     w.namespace,
		ClusterName:   w.clusterName,
		ComponentName: w.componentName,
		InstanceName:  w.instanceName,
		ClientType:    w.clientType,
		Endpoints:     make([]ConnectEndpointJSON, 0, len(targetServices)),
		Accounts:      make([]ConnectAccountJSON, 0),
	}

	portForwardTarget := ""
	portForwardPort := ""
	serviceKind := ""
	nodeExternalAddress := ""
	nodeInternalAddress := ""
	for _, svc := range targetServices {
		if svc.Spec.Type == corev1.ServiceTypeNodePort && nodeExternalAddress == "" && nodeInternalAddress == "" {
			nodeExternalAddress, nodeInternalAddress = w.findReadyNodeAddress(ctx)
		}
		endpoint := buildEndpointJSON(w.namespace, w.instanceName, svc, nodeExternalAddress, nodeInternalAddress)
		resp.Endpoints = append(resp.Endpoints, endpoint)
		if portForwardPort == "" && len(svc.Spec.Ports) > 0 {
			portForwardPort = strconv.Itoa(int(svc.Spec.Ports[0].Port))
			if svc.Spec.ClusterIP != corev1.ClusterIPNone {
				portForwardTarget = fmt.Sprintf("service/%s", svc.Name)
			} else {
				portForwardTarget = w.instanceName
			}
		}
	}

	keys := make([]string, 0, len(componentRefs))
	for key := range componentRefs {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		ref := componentRefs[key]
		if ref.componentDef == "" {
			continue
		}
		componentDef, err := w.kbClient.AppsV1().ComponentDefinitions().Get(ctx, ref.componentDef, metav1.GetOptions{})
		if err != nil {
			continue
		}
		if serviceKind == "" {
			serviceKind = componentDef.Spec.ServiceKind
		}
		for _, account := range componentDef.Spec.SystemAccounts {
			if !account.InitAccount {
				continue
			}
			secretName := constant.GenerateAccountSecretName(w.clusterName, ref.componentName, account.Name)
			passwd := ""
			secret, err := w.kubeClient.CoreV1().Secrets(w.namespace).Get(ctx, secretName, metav1.GetOptions{})
			if err == nil {
				passwd = string(secret.Data["password"])
			}
			resp.Accounts = append(resp.Accounts, ConnectAccountJSON{
				ComponentName: ref.componentName,
				ShardingName:  ref.shardingName,
				SecretName:    secretName,
				Username:      account.Name,
				Password:      passwd,
				PasswordKey:   "<password>",
			})
		}
	}

	resp.ServiceKind = serviceKind
	if portForwardTarget != "" && portForwardPort != "" {
		resp.PortForwardCommand = fmt.Sprintf(
			"kubectl port-forward -n %s %s %s:%s",
			w.namespace,
			portForwardTarget,
			portForwardPort,
			portForwardPort,
		)
	}
	resp.ClientExample = buildClientExample(w.clientType)

	// sort.Slice(resp.Endpoints, func(i, j int) bool {
	// 	if resp.Endpoints[i].ComponentName == resp.Endpoints[j].ComponentName {
	// 		return resp.Endpoints[i].ServiceName < resp.Endpoints[j].ServiceName
	// 	}
	// 	return resp.Endpoints[i].ComponentName < resp.Endpoints[j].ComponentName
	// })
	// sort.Slice(resp.Accounts, func(i, j int) bool {
	// 	if resp.Accounts[i].ComponentName == resp.Accounts[j].ComponentName {
	// 		return resp.Accounts[i].SecretName < resp.Accounts[j].SecretName
	// 	}
	// 	return resp.Accounts[i].ComponentName < resp.Accounts[j].ComponentName
	// })

	return resp, nil
}

func (w *ConnectWrapper) listClusterServices(ctx context.Context) ([]corev1.Service, error) {
	selector := labels.SelectorFromSet(constant.GetClusterLabels(w.clusterName)).String()
	serviceList, err := w.kubeClient.CoreV1().Services(w.namespace).List(ctx, metav1.ListOptions{LabelSelector: selector})
	if err != nil {
		return nil, err
	}
	return serviceList.Items, nil
}

func (w *ConnectWrapper) selectServices(clusterObj *kbappsv1.Cluster, services []corev1.Service) []corev1.Service {
	if w.instanceName != "" {
		return filterServices(services, func(svc corev1.Service) bool {
			return svc.Spec.Selector[constant.KBAppPodNameLabelKey] == w.instanceName ||
				(serviceComponentName(svc) == w.componentName && svc.Spec.ClusterIP == corev1.ClusterIPNone)
		})
	}

	if w.componentName != "" {
		return filterServices(services, func(svc corev1.Service) bool {
			compName := serviceComponentName(svc)
			shardingName := serviceShardingName(svc)
			if compName == w.componentName || shardingName == w.componentName {
				return true
			}
			for _, clusterSvc := range clusterObj.Spec.Services {
				if clusterSvc.ComponentSelector != w.componentName {
					continue
				}
				if svc.Name == constant.GenerateClusterServiceName(w.clusterName, clusterSvc.ServiceName) {
					return true
				}
			}
			return false
		})
	}

	if len(clusterObj.Spec.Services) > 0 {
		clusterServiceNames := make(map[string]struct{}, len(clusterObj.Spec.Services))
		for _, svc := range clusterObj.Spec.Services {
			clusterServiceNames[constant.GenerateClusterServiceName(w.clusterName, svc.ServiceName)] = struct{}{}
		}
		selected := filterServices(services, func(svc corev1.Service) bool {
			// _, ok := clusterServiceNames[svc.Name]
			return true
			// return ok
		})
		if len(selected) > 0 {
			return selected
		}
	}

	return services
}

func (w *ConnectWrapper) resolveComponentRefs(ctx context.Context, clusterObj *kbappsv1.Cluster, services []corev1.Service) (map[string]componentRef, error) {
	refs := make(map[string]componentRef)

	addRef := func(componentName, shardingName, componentDef string) {
		if componentName == "" {
			return
		}
		if componentDef == "" {
			for _, svc := range services {
				if serviceComponentName(svc) == componentName {
					componentDef = svc.Labels[constant.AppComponentLabelKey]
					if componentDef != "" {
						break
					}
				}
			}
		}
		refs[componentName] = componentRef{
			componentName: componentName,
			shardingName:  shardingName,
			componentDef:  componentDef,
		}
	}

	for _, svc := range services {
		componentName := serviceComponentName(svc)
		if componentName == "" {
			continue
		}
		if _, ok := refs[componentName]; ok {
			continue
		}
		if compSpec := clusterObj.Spec.GetComponentByName(componentName); compSpec != nil {
			addRef(componentName, "", compSpec.ComponentDef)
			continue
		}
		if shardingName := serviceShardingName(svc); shardingName != "" {
			componentDef, err := w.resolveShardingComponentDef(ctx, clusterObj, shardingName)
			if err != nil {
				return nil, err
			}
			addRef(componentName, shardingName, componentDef)
		}
	}

	if len(refs) == 0 && w.componentName != "" {
		if compSpec := clusterObj.Spec.GetComponentByName(w.componentName); compSpec != nil {
			addRef(w.componentName, "", compSpec.ComponentDef)
			return refs, nil
		}
		if shardingSpec := clusterObj.Spec.GetShardingByName(w.componentName); shardingSpec != nil {
			componentDef, err := w.resolveShardingComponentDef(ctx, clusterObj, w.componentName)
			if err != nil {
				return nil, err
			}
			addRef(w.componentName, w.componentName, componentDef)
		}
	}

	return refs, nil
}

func (w *ConnectWrapper) resolveShardingComponentDef(ctx context.Context, clusterObj *kbappsv1.Cluster, shardingName string) (string, error) {
	shardingSpec := clusterObj.Spec.GetShardingByName(shardingName)
	if shardingSpec == nil {
		return "", nil
	}
	if shardingSpec.Template.ComponentDef != "" {
		return shardingSpec.Template.ComponentDef, nil
	}
	if shardingSpec.ShardingDef == "" {
		return "", nil
	}
	shardingDef, err := w.kbClient.AppsV1().ShardingDefinitions().Get(ctx, shardingSpec.ShardingDef, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	return shardingDef.Spec.Template.CompDef, nil
}

func (w *ConnectWrapper) findReadyNodeAddress(ctx context.Context) (string, string) {
	nodeList, err := w.kubeClient.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return "", ""
	}
	for _, node := range nodeList.Items {
		if !isNodeReady(node) {
			continue
		}
		internal := ""
		external := ""
		for _, addr := range node.Status.Addresses {
			switch addr.Type {
			case corev1.NodeExternalIP:
				if external == "" {
					external = addr.Address
				}
			case corev1.NodeInternalIP:
				if internal == "" {
					internal = addr.Address
				}
			}
		}
		return external, internal
	}
	return "", ""
}

func isNodeReady(node corev1.Node) bool {
	for _, condition := range node.Status.Conditions {
		if condition.Type == corev1.NodeReady && condition.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}

func buildEndpointJSON(namespace, instanceName string, svc corev1.Service, nodeExternal, nodeInternal string) ConnectEndpointJSON {
	internalHost := fmt.Sprintf("%s.%s.svc.cluster.local", svc.Name, namespace)
	if svc.Spec.ClusterIP == corev1.ClusterIPNone {
		podName := instanceName
		if podName == "" {
			podName = "<podName>"
		}
		internalHost = fmt.Sprintf("%s.%s.%s.svc.cluster.local", podName, svc.Name, namespace)
	}

	external := ""
	switch svc.Spec.Type {
	case corev1.ServiceTypeLoadBalancer:
		for _, ingress := range svc.Status.LoadBalancer.Ingress {
			if ingress.IP != "" {
				external = ingress.IP
				break
			}
			if ingress.Hostname != "" {
				external = ingress.Hostname
				break
			}
		}
	case corev1.ServiceTypeExternalName:
		external = svc.Spec.ExternalName
	case corev1.ServiceTypeNodePort:
		if nodeExternal != "" {
			external = nodeExternal
		} else {
			external = nodeInternal
		}
	default:
		if len(svc.Spec.ExternalIPs) > 0 {
			external = svc.Spec.ExternalIPs[0]
		}
	}

	ports := make([]ConnectPortJSON, 0, len(svc.Spec.Ports))
	for _, port := range svc.Spec.Ports {
		ports = append(ports, ConnectPortJSON{
			Name:     port.Name,
			Port:     port.Port,
			NodePort: port.NodePort,
		})
	}

	return ConnectEndpointJSON{
		ComponentName: serviceComponentName(svc),
		ShardingName:  serviceShardingName(svc),
		ServiceName:   svc.Name,
		Type:          string(svc.Spec.Type),
		Internal:      internalHost,
		External:      external,
		Ports:         ports,
	}
}

func serviceComponentName(svc corev1.Service) string {
	if svc.Annotations[constant.KBAppComponentLabelKey] != "" {
		return svc.Annotations[constant.KBAppComponentLabelKey]
	}
	if svc.Labels[constant.KBAppComponentLabelKey] != "" {
		return svc.Labels[constant.KBAppComponentLabelKey]
	}
	return svc.Spec.Selector[constant.KBAppComponentLabelKey]
}

func serviceShardingName(svc corev1.Service) string {
	if svc.Annotations[constant.KBAppShardingNameLabelKey] != "" {
		return svc.Annotations[constant.KBAppShardingNameLabelKey]
	}
	return svc.Labels[constant.KBAppShardingNameLabelKey]
}

func filterServices(services []corev1.Service, match func(corev1.Service) bool) []corev1.Service {
	result := make([]corev1.Service, 0, len(services))
	for _, svc := range services {
		if !match(svc) {
			continue
		}
		result = append(result, svc)
	}
	return result
}

func buildClientExample(clientType string) string {
	switch clientType {
	case "python":
		return "host='<HOST>' port=<PORT> user='<USERNAME>' password='<PASSWORD>'"
	case "java":
		return "jdbc:<SERVICE>://<HOST>:<PORT> user=<USERNAME> password=<PASSWORD>"
	case "node.js", "node", "javascript":
		return "host: '<HOST>', port: <PORT>, user: '<USERNAME>', password: '<PASSWORD>'"
	case "go":
		return `dsn := "<USERNAME>:<PASSWORD>@tcp(<HOST>:<PORT>)/"`
	case "cli", "":
		return "<CLIENT> -h <HOST> -P <PORT> -u <USERNAME> -p<PASSWORD>"
	default:
		return "host=<HOST> port=<PORT> user=<USERNAME> password=<PASSWORD>"
	}
}
