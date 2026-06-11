package helm

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"strings"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/downloader"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/release"
	"k8s.io/cli-runtime/pkg/resource"
)

func RunInstall(ctx context.Context, logger *log.Logger, settings *cli.EnvSettings, releaseName string, chartRef string, chartVersion string, releaseValues map[string]interface{}) (*release.Release, error) {

	actionConfig, err := initActionConfig(settings, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to init action config: %w", err)
	}

	installClient := action.NewInstall(actionConfig)

	installClient.DryRun = true
	installClient.ReleaseName = releaseName
	installClient.Namespace = settings.Namespace()
	installClient.Version = chartVersion
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

	providers := getter.All(settings)

	charter, err := loader.Load(chartPath)
	if err != nil {
		return nil, err
	}

	// Check chart dependencies to make sure all are present in /charts
	if chartDependencies := charter.Metadata.Dependencies; chartDependencies != nil {
		if err := action.CheckDependencies(charter, chartDependencies); err != nil {
			err = fmt.Errorf("failed to check chart dependencies: %w", err)
			if !installClient.DependencyUpdate {
				return nil, err
			}

			manager := &downloader.Manager{
				Out:              logger.Writer(),
				ChartPath:        chartPath,
				Keyring:          installClient.ChartPathOptions.Keyring,
				SkipUpdate:       false,
				Getters:          providers,
				RepositoryConfig: settings.RepositoryConfig,
				RepositoryCache:  settings.RepositoryCache,
				Debug:            settings.Debug,
				RegistryClient:   installClient.GetRegistryClient(),
			}
			if err := manager.Update(); err != nil {
				return nil, err
			}
			// Reload the chart with the updated Chart.lock file.
			if charter, err = loader.Load(chartPath); err != nil {
				return nil, fmt.Errorf("failed to reload chart after repo update: %w", err)
			}
		}
	}

	rel, err := installClient.RunWithContext(ctx, charter, releaseValues)
	if err != nil {
		return nil, fmt.Errorf("failed to run install: %w", err)
	}

	logger.Printf("release created")

	return rel, nil
}

func TryRunInstallGetClusterManifest(ctx context.Context, logger *log.Logger, settings *cli.EnvSettings, releaseName string, chartRef string, chartVersion string, releaseValues map[string]interface{}) (*resource.Info, error) {
	rel, err := RunInstall(ctx, logger, settings, releaseName, chartRef, chartVersion, releaseValues)
	if err != nil {
		return nil, err
	}
	manifest := rel.Manifest
	actionConfig, err := initActionConfig(settings, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to init action config: %w", err)
	}
	current, err := actionConfig.KubeClient.Build(bytes.NewBufferString(manifest), false)
	if err != nil {
		return nil, fmt.Errorf("current")
	}
	rs := current.Filter(func(i *resource.Info) bool {
		return i.Mapping.GroupVersionKind.Kind == "Cluster"
	})
	if len(rs) > 0 {
		return rs[0], err
	}
	return nil, fmt.Errorf("not found cluster manifest")

}

// CheckReleaseExists 检查指定命名空间中的 release 是否存在
func CheckReleaseExists(logger *log.Logger, namespace string, releaseName string) (bool, error) {
	settings := cli.New()
	settings.SetNamespace(namespace)

	actionConfig, err := initActionConfig(settings, logger)
	if err != nil {
		return false, fmt.Errorf("failed to init action config: %w", err)
	}

	// 使用 Get 操作检查 release 是否存在
	getClient := action.NewGet(actionConfig)
	_, err = getClient.Run(releaseName)
	if err != nil {
		// 如果返回的 error 包含 "not found"，说明 release 不存在
		if strings.Contains(err.Error(), "not found") {
			return false, nil
		}
		return false, fmt.Errorf("failed to get release: %w", err)
	}

	return true, nil
}
