package tool

import (
	"fmt"
	"os"
	"time"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/cli/values"
	"helm.sh/helm/v3/pkg/getter"
	"k8s.io/helm/pkg/strvals"
)

type Helm struct {
	config      *action.Configuration
	envSettings *cli.EnvSettings
	args        map[string]interface{}
}

func NewHelm(envSettings *cli.EnvSettings, args map[string]interface{}) (*Helm, error) {
	helm := &Helm{envSettings: envSettings, args: args}
	config := new(action.Configuration)
	if err := config.Init(envSettings.RESTClientGetter(), envSettings.Namespace(), os.Getenv("HELM_DRIVER"), func(format string, v ...interface{}) {
		LogInfo(fmt.Sprintf(format, v))
	}); err != nil {
		return nil, err
	}
	helm.config = config
	return helm, nil
}

func (h Helm) Install(namespace, chart, release, valuesFile string) error {
	LogInfo(fmt.Sprintf("Execute helm install. namespace: %s, release: %s chart: %s", namespace, release, chart))
	client := action.NewInstall(h.config)
	client.Namespace = namespace
	client.ReleaseName = release
	client.Wait = true
	// default timeout duration
	// ref: https://helm.sh/docs/helm/helm_install
	client.Timeout = 5 * time.Minute

	cp, err := client.ChartPathOptions.LocateChart(chart, h.envSettings)
	if err != nil {
		LogError(fmt.Sprintf("Error LocateChart: %v", err))
		return err
	}

	p := getter.All(h.envSettings)
	valueOpts := &values.Options{}
	if valuesFile != "" {
		valueOpts.ValueFiles = append(valueOpts.ValueFiles, valuesFile)
	}
	vals, err := valueOpts.MergeValues(p)
	if err != nil {
		LogError(fmt.Sprintf("Error MergeValues: %v", err))
		return err
	}

	if val, ok := h.args["set"]; ok {
		if err := strvals.ParseInto(fmt.Sprintf("%v", val), vals); err != nil {
			LogError(fmt.Sprintf("Error parsing --set values: %v", err))
			return err
		}
	}

	if val, ok := h.args["set-file"]; ok {
		if err := strvals.ParseInto(fmt.Sprintf("%v", val), vals); err != nil {
			LogError(fmt.Sprintf("Error parsing --set-file values: %v", err))
			return err
		}
	}

	if val, ok := h.args["set-string"]; ok {
		if err := strvals.ParseInto(fmt.Sprintf("%v", val), vals); err != nil {
			LogError(fmt.Sprintf("Error parsing --set-string values: %v", err))
			return err
		}
	}

	c, err := loader.Load(cp)
	if err != nil {
		LogError(fmt.Sprintf("Error loading chart path: %v", err))
		return err
	}

	// TODO: support other options if required
	_, err = client.Run(c, vals)
	if err != nil {
		LogError(fmt.Sprintf("Error running chart install: %v", err))
		return err
	}

	LogInfo("Helm install complete")
	return nil
}

func (h Helm) Test(namespace, release string) error {
	LogInfo(fmt.Sprintf("Execute helm test. namespace: %s, release: %s, args: %+v", namespace, release, h.args))
	client := action.NewReleaseTesting(h.config)
	client.Namespace = namespace

	// TODO: support filter and timeout options if required
	_, err := client.Run(release)
	if err != nil {
		LogError(fmt.Sprintf("Execute helm test. error %v", err))
		return err
	}

	LogInfo("Helm test complete")
	return nil
}

func (h Helm) Uninstall(namespace, release string) error {
	LogInfo(fmt.Sprintf("Execute helm uninstall. namespace: %s, release: %s", namespace, release))
	client := action.NewUninstall(h.config)
	// TODO: support other options if required
	_, err := client.Run(release)

	if err != nil {
		LogError(fmt.Sprintf("Error from helm uninstall : %v", err))
		return err
	}

	LogInfo("Delete release complete")
	return nil
}

func (h Helm) Upgrade(namespace, chart, release string) error {
	LogInfo(fmt.Sprintf("Execute helm upgrade. namespace: %s, release: %s chart: %s", namespace, release, chart))
	client := action.NewUpgrade(h.config)
	client.Namespace = namespace
	client.ReuseValues = true
	client.Wait = true

	cp, err := client.ChartPathOptions.LocateChart(chart, h.envSettings)
	if err != nil {
		LogError(fmt.Sprintf("Error LocateChart: %v", err))
		return err
	}

	p := getter.All(h.envSettings)
	valueOpts := &values.Options{}
	vals, err := valueOpts.MergeValues(p)
	if err != nil {
		LogError(fmt.Sprintf("Error MergeValues: %v", err))
		return err
	}

	c, err := loader.Load(cp)
	if err != nil {
		LogError(fmt.Sprintf("Error loading chart path: %v", err))
		return err
	}

	// TODO: support other options if required
	_, err = client.Run(release, c, vals)
	if err != nil {
		LogError(fmt.Sprintf("Error running chart upgrade: %v", err))
		return err
	}

	LogInfo("Helm upgrade complete")
	return nil
}
