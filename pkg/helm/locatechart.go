package helm

import (
	"fmt"
	"log"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
)

func LocateChart(settings *cli.EnvSettings, chartRef, version string) (*chart.Chart, error) {
	actionConfig, err := initActionConfig(settings, log.Default())
	if err != nil {
		return nil, fmt.Errorf("failed to init action config: %w", err)
	}

	installClient := action.NewInstall(actionConfig)

	installClient.DryRun = true
	installClient.Version = version
	installClient.RepoURL = KUBE_CLUSETER_REPO

	registryClient, err := newRegistryClient(
		settings,
		installClient.CertFile,
		installClient.KeyFile,
		installClient.CaFile,
		false,
		installClient.PlainHTTP)
	if err != nil {
		return nil, fmt.Errorf("failed to created registry client: %w", err)
	}
	installClient.SetRegistryClient(registryClient)

	chartPath, err := installClient.ChartPathOptions.LocateChart(chartRef, settings)
	if err != nil {
		return nil, err
	}

	charter, err := loader.Load(chartPath)
	if err != nil {
		return nil, err
	}
	return charter, nil
}
