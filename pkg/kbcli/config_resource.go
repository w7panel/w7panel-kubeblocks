/*
Copyright (C) 2022-2026 ApeCloud Co., Ltd

This file is part of KubeBlocks project

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package kbcli

import (
	"context"
	"fmt"
	"kbcli-helper/pkg/format"
	"sort"
	"strings"

	appsv1 "github.com/apecloud/kubeblocks/apis/apps/v1"
	parametersv1alpha1 "github.com/apecloud/kubeblocks/apis/parameters/v1alpha1"
	"github.com/apecloud/kubeblocks/pkg/client/clientset/versioned"
	"github.com/apecloud/kubeblocks/pkg/configuration/openapi"
	cfgutil "github.com/apecloud/kubeblocks/pkg/configuration/util"
	intctrlutil "github.com/apecloud/kubeblocks/pkg/controllerutil"
	apiext "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ConfigObjectsWrapper struct {
	namespace   string
	clusterName string
	components  []string

	rctxMap    map[string]*ReconfigureContext
	readConfig func(ConfigMapName, namespace, key string) (string, error)
}

type ConfigObjectsJSON struct {
	Namespace   string                `json:"namespace"`
	ClusterName string                `json:"clusterName"`
	Components  []ComponentConfigJSON `json:"components"`
}

type ComponentConfigJSON struct {
	ComponentName string                                  `json:"componentName"`
	Templates     []ConfigTemplateJSON                    `json:"templates"`
	Render        *parametersv1alpha1.ParamConfigRenderer `json:"render"`
}

type ConfigTemplateJSON struct {
	FileName         string                               `json:"fileName"`
	ConfigSpec       string                               `json:"configSpec,omitempty"`
	ComponentName    string                               `json:"componentName"`
	ClusterName      string                               `json:"clusterName"`
	SchemaAvailable  bool                                 `json:"schemaAvailable"`
	Message          string                               `json:"message,omitempty"`
	Parameters       []ParameterSchemaJSON                `json:"parameters"`
	FileFormatConfig *parametersv1alpha1.FileFormatConfig `json:"fileFormatConfig,omitempty"`
	ConfigMapName    string                               `json:"confitMapName,omitempty"`
	ConfigData       map[string]string                    `json:"configData,omitempty"`
}

type ParameterSchemaJSON struct {
	Name          string   `json:"name"`
	AllowedValues string   `json:"allowedValues"`
	DefaultValue  string   `json:"defaultValue,omitempty"`
	CurrentValue  string   `json:"currentValue,omitempty"`
	Scope         string   `json:"scope"`
	Dynamic       bool     `json:"dynamic"`
	Type          string   `json:"type"`
	Description   string   `json:"description,omitempty"`
	Enum          []string `json:"enum,omitempty"`
	Minimum       string   `json:"minimum,omitempty"`
	Maximum       string   `json:"maximum,omitempty"`
}

func (r *ConfigObjectsWrapper) ToJson() (*ConfigObjectsJSON, error) {
	resp := &ConfigObjectsJSON{
		Namespace:   r.namespace,
		ClusterName: r.clusterName,
		Components:  make([]ComponentConfigJSON, 0, len(r.components)),
	}

	for _, component := range r.components {
		rctx, ok := r.rctxMap[component]
		if !ok || rctx == nil {
			continue
		}
		item, err := r.buildComponentExplainConfigure(rctx)
		if err != nil {
			return nil, err
		}
		resp.Components = append(resp.Components, *item)
	}
	return resp, nil
}

type parameterSchema struct {
	name         string
	valueType    string
	miniNum      string
	maxiNum      string
	enum         []string
	description  string
	scope        string
	dynamic      bool
	defaultValue string
}

func (pt *parameterSchema) enumFormatter(maxFieldLength int) string {
	if len(pt.enum) == 0 {
		return ""
	}
	v := strings.Join(pt.enum, ",")
	if maxFieldLength > 0 && len(v) > maxFieldLength {
		v = v[:maxFieldLength] + "..."
	}
	return v
}

func (pt *parameterSchema) rangeFormatter() string {
	const (
		r          = "-"
		rangeBegin = "["
		rangeEnd   = "]"
	)

	if len(pt.maxiNum) == 0 && len(pt.miniNum) == 0 {
		return ""
	}

	v := rangeBegin
	if len(pt.miniNum) != 0 {
		v += pt.miniNum
	}
	if len(pt.maxiNum) != 0 {
		v += r
		v += pt.maxiNum
	} else if len(v) != 0 {
		v += r
	}
	v += rangeEnd
	return v
}

func GetCluster(clientSet versioned.Interface, ns, clusterName string) (*appsv1.Cluster, error) {
	clusterObj, err := clientSet.AppsV1().Clusters(ns).Get(context.TODO(), clusterName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return clusterObj, nil
}

func New(clusterName string, namespace string, components ...string) (*ConfigObjectsWrapper, error) {
	clientSet, err := getClient()
	if err != nil {
		return nil, err
	}
	clusterObj, err := GetCluster(clientSet, namespace, clusterName)
	if err != nil {
		return nil, err
	}
	if len(components) == 0 {
		components = getComponentNames(clusterObj)
	}

	rctxAsMap := make(map[string]*ReconfigureContext)
	for _, compName := range components {
		rctx, err := generateReconfigureContext(context.TODO(), clientSet, clusterName, compName, namespace)
		if err != nil {
			return nil, err
		}
		rctxAsMap[rctx.CompName] = rctx
	}

	return &ConfigObjectsWrapper{
		namespace:   namespace,
		components:  components,
		clusterName: clusterName,
		rctxMap:     rctxAsMap,
		readConfig:  readConfig,
	}, nil
}

func (r *ConfigObjectsWrapper) buildComponentExplainConfigure(rctx *ReconfigureContext) (*ComponentConfigJSON, error) {
	componentConfig := &ComponentConfigJSON{
		ComponentName: rctx.CompName,
		Templates:     make([]ConfigTemplateJSON, 0, len(rctx.ParametersDefs)),
	}

	for _, pd := range rctx.ParametersDefs {
		template, err := r.buildExplainConfigure(rctx, &pd.Spec)
		if err != nil {
			return nil, err
		}
		componentConfig.Templates = append(componentConfig.Templates, *template)
	}
	return componentConfig, nil
}

func (r *ConfigObjectsWrapper) buildExplainConfigure(rctx *ReconfigureContext, pdSpec *parametersv1alpha1.ParametersDefinitionSpec) (*ConfigTemplateJSON, error) {
	template := &ConfigTemplateJSON{
		FileName:      pdSpec.FileName,
		ComponentName: rctx.CompName,
		ClusterName:   r.clusterName,
		Parameters:    make([]ParameterSchemaJSON, 0),
	}

	if rctx.ConfigRender != nil {
		config := intctrlutil.GetComponentConfigDescription(&rctx.ConfigRender.Spec, pdSpec.FileName)
		if config != nil {
			template.ConfigSpec = config.TemplateName
			template.FileFormatConfig = config.FileFormatConfig
			template.ConfigMapName = fmt.Sprintf("%s-%s-%s", r.clusterName, rctx.CompName, config.TemplateName)
			if r.readConfig != nil {
				data, err := r.readConfig(template.ConfigMapName, r.namespace, config.Name)
				if err != nil {
					return nil, err
				}
				formatD := template.FileFormatConfig.Format
				sectionName := ""
				if formatD == "ini" {
					sectionName = template.FileFormatConfig.IniConfig.SectionName
				}
				data2, err := format.Format([]byte(data), format.CfgFileFormat(formatD), sectionName)
				if err != nil {
					return nil, err
				}
				template.ConfigData = data2

			}
		}
	}

	if pdSpec.ParametersSchema == nil {
		template.Message = fmt.Sprintf(notConfigSchemaPrompt, pdSpec.FileName)
		return template, nil
	}

	schema := pdSpec.ParametersSchema.DeepCopy()
	if schema.SchemaInJSON == nil {
		if schema.CUE == "" {
			template.Message = fmt.Sprintf(notConfigSchemaPrompt, pdSpec.FileName)
			return template, nil
		}
		apiSchema, err := openapi.GenerateOpenAPISchema(schema.CUE, schema.TopLevelKey)
		if err != nil {
			return nil, err
		}
		if apiSchema == nil {
			template.Message = cue2openAPISchemaFailedPrompt
			return template, nil
		}
		schema.SchemaInJSON = apiSchema
	}

	params, err := r.buildConfigConstraint(schema.SchemaInJSON,
		cfgutil.NewSet(pdSpec.StaticParameters...),
		cfgutil.NewSet(pdSpec.DynamicParameters...),
		cfgutil.NewSet(pdSpec.ImmutableParameters...), template.ConfigData)
	if err != nil {
		return nil, err
	}
	template.SchemaAvailable = true
	template.Parameters = params
	return template, nil
}

// func (r *ConfigObjectsWrapper) hasSpecificParam() bool {
// 	return len(r.paramName) != 0
// }

//	func (r *ConfigObjectsWrapper) isSpecificParam(paramName string) bool {
//		return r.paramName == paramName
//	}
func (r *ConfigObjectsWrapper) buildConfigConstraint(schema *apiext.JSONSchemaProps, staticParameters, dynamicParameters, immutableParameters *cfgutil.Sets, configData map[string]string) ([]ParameterSchemaJSON, error) {
	var (
		spec   = schema.Properties[openapi.DefaultSchemaName]
		params = make([]*parameterSchema, 0)
	)

	for key, property := range openapi.FlattenSchema(spec).Properties {
		if property.Type == openapi.SchemaStructType {
			continue
		}
		// if r.hasSpecificParam() && !r.isSpecificParam(key) {
		// 	continue
		// }

		pt, err := generateParameterSchema(key, property)
		if err != nil {
			return nil, err
		}
		pt.scope = "Global"
		pt.dynamic = isDynamicType(pt, staticParameters, dynamicParameters, immutableParameters)

		params = append(params, pt)
	}

	sort.SliceStable(params, func(i, j int) bool {
		x1 := params[i]
		x2 := params[j]
		return strings.Compare(x1.name, x2.name) < 0
	})

	resp := make([]ParameterSchemaJSON, 0, len(params))
	for _, pt := range params {
		resp = append(resp, ParameterSchemaJSON{
			Name:          pt.name,
			AllowedValues: getAllowedValues(pt, 0),
			DefaultValue:  pt.defaultValue,
			Scope:         pt.scope,
			Dynamic:       pt.dynamic,
			Type:          pt.valueType,
			Description:   pt.description,
			Enum:          pt.enum,
			Minimum:       pt.miniNum,
			Maximum:       pt.maxiNum,
		})
	}
	if configData != nil {
		for k, v := range configData {
			for i := range resp {
				if resp[i].Name == k {
					resp[i].CurrentValue = v
				}
			}
		}
	}
	return resp, nil
}

func isDynamicType(pt *parameterSchema, staticParameters, dynamicParameters, immutableParameters *cfgutil.Sets) bool {
	switch {
	case immutableParameters.InArray(pt.name):
		return false
	case staticParameters.InArray(pt.name):
		return false
	case dynamicParameters.InArray(pt.name):
		return true
	case dynamicParameters.Length() == 0 && staticParameters.Length() != 0:
		return true
	case dynamicParameters.Length() != 0 && staticParameters.Length() == 0:
		return false
	default:
		return false
	}
}
