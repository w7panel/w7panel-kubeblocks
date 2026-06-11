package helm

import (
	"testing"

	"helm.sh/helm/v3/pkg/cli"
)

func TestPull(t *testing.T) {

	// logger := log.Default()
	// cli.New()
	err := runPull(cli.New(), "tidb-cluster", "1.0.1")
	if err != nil {
		t.Log(err)
	}
}
