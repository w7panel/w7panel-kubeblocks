package kbcli

import (
	"context"
	"encoding/json"
	"strings"

	appsv1 "github.com/apecloud/kubeblocks/apis/apps/v1"
	apiext "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func generateParameterSchema(paramName string, property apiext.JSONSchemaProps) (*parameterSchema, error) {
	toString := func(v interface{}) (string, error) {
		b, err := json.Marshal(v)
		if err != nil {
			return "", err
		}
		return string(b), nil
	}
	pt := &parameterSchema{
		name:        paramName,
		valueType:   property.Type,
		description: strings.TrimSpace(property.Description),
	}
	if property.Minimum != nil {
		b, err := toString(property.Minimum)
		if err != nil {
			return nil, err
		}
		pt.miniNum = b
	}
	if property.Default != nil {
		b, err := toString(property.Default)
		if err != nil {
			return nil, err
		}
		pt.defaultValue = b
	}
	if property.Format != "" {
		pt.valueType = property.Format
	}
	if property.Maximum != nil {
		b, err := toString(property.Maximum)
		if err != nil {
			return nil, err
		}
		pt.maxiNum = b
	}
	if property.Enum != nil {
		pt.enum = make([]string, len(property.Enum))
		for i, v := range property.Enum {
			b, err := toString(v)
			if err != nil {
				return nil, err
			}
			pt.enum[i] = b
		}
	}
	return pt, nil
}

func getAllowedValues(pt *parameterSchema, maxFieldLength int) string {
	if len(pt.enum) != 0 {
		return pt.enumFormatter(maxFieldLength)
	}
	return pt.rangeFormatter()
}

func getComponentNames(cluster *appsv1.Cluster) []string {
	var components []string
	for _, component := range cluster.Spec.ComponentSpecs {
		components = append(components, component.Name)
	}
	for _, component := range cluster.Spec.Shardings {
		components = append(components, component.Name)
	}
	return components
}

func readConfig(name string, namespace, key string) (string, error) {
	client, err := getDefaultClient()
	if err != nil {
		return "", err
	}

	configMap, err := client.CoreV1().ConfigMaps(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	return (configMap.Data[key]), nil
}
