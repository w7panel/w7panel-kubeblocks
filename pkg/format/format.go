package format

import (
	"fmt"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/clbanning/mxj/v2"
	"github.com/hashicorp/hcl"
	"github.com/magiconair/properties"
	"github.com/subosito/gotenv"
	"gopkg.in/ini.v1"
	"gopkg.in/yaml.v3"
)

func Format(data []byte, format CfgFileFormat, seletion string) (map[string]string, error) {
	switch format {
	case Ini:
		return parseINI(data, seletion)
	case YAML:
		return parseStructured(data, "yaml")
	case JSON:
		return parseStructured(data, "json")
	case XML:
		return parseXML(data)
	case HCL:
		return parseHCL(data)
	case Dotenv:
		return parseDotenv(data)
	case TOML:
		return parseTOML(data)
	case Properties, PropertiesPlus, PropertiesUltra:
		return parseProperties(data)
	case RedisCfg:
		return parseRedis(data)
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

func parseINI(data []byte, seletion string) (map[string]string, error) {
	cfg, err := ini.Load(data)
	if err != nil {
		return nil, err
	}

	result := make(map[string]string)
	for _, section := range cfg.Sections() {
		if section.Name() != seletion {
			continue
		}
		// name := section.Name()
		for _, key := range section.Keys() {
			name := key.Name()
			// if sectionName != ini.DefaultSection {
			// 	name = sectionName + "." + name
			// }
			result[name] = key.Value()
		}
	}
	return result, nil
}

func parseStructured(data []byte, kind string) (map[string]string, error) {
	var parsed interface{}
	switch kind {
	case "yaml":
		if err := yaml.Unmarshal(data, &parsed); err != nil {
			return nil, err
		}
	case "json":
		mv, err := mxj.NewMapJson(data)
		if err != nil {
			return nil, err
		}
		parsed = map[string]interface{}(mv)
	default:
		return nil, fmt.Errorf("unsupported structured kind: %s", kind)
	}

	result := make(map[string]string)
	flattenValue(result, "", parsed)
	return result, nil
}

func parseXML(data []byte) (map[string]string, error) {
	mv, err := mxj.NewMapXml(data)
	if err != nil {
		return nil, err
	}

	result := make(map[string]string)
	flattenValue(result, "", map[string]interface{}(mv))
	return result, nil
}

func parseHCL(data []byte) (map[string]string, error) {
	var parsed map[string]interface{}
	if err := hcl.Unmarshal(data, &parsed); err != nil {
		return nil, err
	}

	result := make(map[string]string)
	flattenValue(result, "", parsed)
	return result, nil
}

func parseDotenv(data []byte) (map[string]string, error) {
	env, err := gotenv.StrictParse(strings.NewReader(string(data)))
	if err != nil {
		return nil, err
	}

	result := make(map[string]string, len(env))
	for k, v := range env {
		result[k] = v
	}
	return result, nil
}

func parseTOML(data []byte) (map[string]string, error) {
	var parsed map[string]interface{}
	if err := toml.Unmarshal(data, &parsed); err != nil {
		return nil, err
	}

	result := make(map[string]string)
	flattenValue(result, "", parsed)
	return result, nil
}

func parseProperties(data []byte) (map[string]string, error) {
	props, err := properties.Load(data, properties.UTF8)
	if err != nil {
		return nil, err
	}
	return props.Map(), nil
}

func parseRedis(data []byte) (map[string]string, error) {
	result := make(map[string]string)
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}

		key := fields[0]
		value := ""
		if len(fields) > 1 {
			value = strings.Join(fields[1:], " ")
		}
		result[key] = value
	}
	return result, nil
}

func flattenValue(result map[string]string, prefix string, value interface{}) {
	switch v := value.(type) {
	case nil:
		result[normalizePath(prefix)] = ""
	case map[string]interface{}:
		for key, item := range v {
			flattenValue(result, joinPath(prefix, key), item)
		}
	case map[interface{}]interface{}:
		for key, item := range v {
			flattenValue(result, joinPath(prefix, fmt.Sprint(key)), item)
		}
	case []interface{}:
		for i, item := range v {
			flattenValue(result, fmt.Sprintf("%s[%d]", normalizeArrayPrefix(prefix), i), item)
		}
	default:
		result[normalizePath(prefix)] = fmt.Sprint(v)
	}
}

func joinPath(prefix, key string) string {
	if prefix == "" {
		return key
	}
	return prefix + "." + key
}

func normalizeArrayPrefix(prefix string) string {
	if prefix == "" {
		return "value"
	}
	return prefix
}

func normalizePath(path string) string {
	if path == "" {
		return "value"
	}
	return path
}
