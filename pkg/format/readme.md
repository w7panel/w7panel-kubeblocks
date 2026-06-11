```
type FileFormatConfig struct {
	// Each format may have its own set of parameters that can be configured.
	// For instance, when using the `ini` format, you can specify the section name.
	//
	// +optional
	FormatterAction `json:",inline"`

	// The config file format. Valid values are `ini`, `xml`, `yaml`, `json`,
	// `hcl`, `dotenv`, `properties` and `toml`. Each format has its own characteristics and use cases.
	//
	// - ini: is a text-based content with a structure and syntax comprising key–value pairs for properties, reference wiki: https://en.wikipedia.org/wiki/INI_file
	// - xml: refers to wiki: https://en.wikipedia.org/wiki/XML
	// - yaml: supports for complex data types and structures.
	// - json: refers to wiki: https://en.wikipedia.org/wiki/JSON
	// - hcl: The HashiCorp Configuration Language (HCL) is a configuration language authored by HashiCorp, reference url: https://www.linode.com/docs/guides/introduction-to-hcl/
	// - dotenv: is a plain text file with simple key–value pairs, reference wiki: https://en.wikipedia.org/wiki/Configuration_file#MS-DOS
	// - properties: a file extension mainly used in Java, reference wiki: https://en.wikipedia.org/wiki/.properties
	// - toml: refers to wiki: https://en.wikipedia.org/wiki/TOML
	// - props-plus: a file extension mainly used in Java, supports CamelCase(e.g: brokerMaxConnectionsPerIp)
	//
	// +kubebuilder:validation:Required
	Format CfgFileFormat `json:"format"`
}
```