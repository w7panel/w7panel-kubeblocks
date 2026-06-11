package helm

import (
	"fmt"
	"log/slog"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
)

func PullChart(settings *cli.EnvSettings, chartRef, chartVersion, destDir string) error {
	actionConfig, err := initActionConfig(settings, nil)
	if err != nil {
		return fmt.Errorf("failed to init action config: %w", err)
	}
	pullClient := action.NewPullWithOpts(action.WithConfig(actionConfig))
	pullClient.DestDir = destDir
	pullClient.Settings = settings
	pullClient.Version = chartVersion
	pullClient.RepoURL = KUBE_CLUSETER_REPO
	pullClient.Untar = true

	registryClient, err := newRegistryClient(
		settings,
		pullClient.CertFile,
		pullClient.KeyFile,
		pullClient.CaFile,
		pullClient.InsecureSkipTLSverify,
		pullClient.PlainHTTP)
	if err != nil {
		return fmt.Errorf("failed to created registry client: %w", err)
	}
	actionConfig.RegistryClient = registryClient

	result, err := pullClient.Run(chartRef)
	if err != nil {
		return fmt.Errorf("failed to pull chart: %w", err)
	}
	slog.Info("Chart pulled successfully", "result", result)

	return nil
}

func runPull(settings *cli.EnvSettings, chartRef, chartVersion string) error {
	return PullChart(settings, chartRef, chartVersion, "./")
}
