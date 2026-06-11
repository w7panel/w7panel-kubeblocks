package helm

import (
	"context"
	"log"
	"log/slog"
	"testing"

	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/cli/values"
	"helm.sh/helm/v3/pkg/getter"
)

func TestInstall(t *testing.T) {

	// logger := log.Default()
	// cli.New()
	settings := cli.New()
	val, err := getVal(settings)
	if err != nil {
		t.Log(err)
		return
	}
	release, err := RunInstall(context.Background(), log.Default(), cli.New(), "hello", "tidb-cluster", "1.0.1", val)
	if err != nil {
		t.Log(err)
		return
	}
	t.Log(release)

}

func TestInstall2(t *testing.T) {

	// logger := log.Default()
	// cli.New()
	settings := cli.New()
	val, err := getVal(settings)
	if err != nil {
		t.Log(err)
		return
	}
	info, err := TryRunInstallGetClusterManifest(context.Background(), log.Default(), cli.New(), "hello", "tidb-cluster", "1.0.1", val)
	if err != nil {
		t.Log(err)
		return
	}
	t.Log(info.Object)

}

func getVal(settings *cli.EnvSettings) (map[string]interface{}, error) {
	optValues := &values.Options{}
	optValues.Values = []string{"pd.replicas=4"}
	// for _, val := range helmOp.set {
	// 	optValues.StringValues = append(optValues.StringValues, val)
	// }
	// for _, val := range helmOp.setJson {
	// 	optValues.JSONValues = append(optValues.JSONValues, val)
	// }
	slog.Info("abcdefg merge values", "vals", optValues.JSONValues)
	provider := getter.All(settings)
	vals, err := optValues.MergeValues(provider)

	return vals, err
}
