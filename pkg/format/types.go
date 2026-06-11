package format

type CfgFileFormat string

const (
	Ini             CfgFileFormat = "ini"
	YAML            CfgFileFormat = "yaml"
	JSON            CfgFileFormat = "json"
	XML             CfgFileFormat = "xml"
	HCL             CfgFileFormat = "hcl"
	Dotenv          CfgFileFormat = "dotenv"
	TOML            CfgFileFormat = "toml"
	Properties      CfgFileFormat = "properties"
	RedisCfg        CfgFileFormat = "redis"
	PropertiesPlus  CfgFileFormat = "props-plus"
	PropertiesUltra CfgFileFormat = "props-ultra"
)
