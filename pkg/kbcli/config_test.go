package kbcli

import (
	"testing"

	"helm.sh/helm/v3/pkg/cli"
)

func Test_Config(t *testing.T) {

	helmApi := cli.New()
	restConfig, err := helmApi.RESTClientGetter().ToRESTConfig()
	if err != nil {
		t.Skipf("skipping: failed to build rest config: %v", err)
	}
	client := GetClientFromRestConfig(restConfig)
	// NewConfigWrapper(client, "default", "mysql1", "mysql", "mysql", "mysql-config.cue", nil)
	wrapper, err := NewConfigWrapper(client, "default", "mysql57", "mysql", "", "", nil)
	if err != nil {
		t.Skipf("skipping: test cluster data not available: %v", err)
	}
	specName := wrapper.ConfigSpecName()
	t.Log(specName)

}

func TestLoadConfigParams(t *testing.T) {

	war, err := New("mysql57", "default")
	if err != nil {
		t.Skipf("skipping: test cluster data not available: %v", err)
	}
	json, err := war.ToJson()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(json)
}
