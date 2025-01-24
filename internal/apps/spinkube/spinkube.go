package spinkube

import (
	"fmt"
	"github.com/ksctl/ka/internal/apps"
	"github.com/ksctl/ka/internal/apps/kwasm"
	"github.com/ksctl/ksctl/pkg/apps/stack"
	"github.com/ksctl/ksctl/pkg/helm"
	"github.com/ksctl/ksctl/pkg/k8s"
	"strings"

	"github.com/ksctl/ksctl/pkg/poller"

	"github.com/ksctl/ksctl/pkg/utilities"
)

func getSpinkubeComponentOverridings(p stack.ComponentOverrides) (version *string) {
	if p == nil {
		return nil
	}

	for k, v := range p {
		switch k {
		case "version":
			if v, ok := v.(string); ok {
				version = utilities.Ptr(v)
			}
		}
	}
	return
}

func GetSpinKubeStackSpecificKwasmOverrides(params stack.ComponentOverrides) error {
	releases, err := poller.GetSharedPoller().Get("spinkube", "containerd-shim-spin")
	if err != nil {
		return err
	}
	nodeInstallerOCI := "ghcr.io/spinkube/containerd-shim-spin/node-installer:" + releases[0]

	if params == nil {
		params = stack.ComponentOverrides{}
	}
	if _, ok := params[kwasm.OperatorChartOverridingsKey]; !ok {
		params[kwasm.OperatorChartOverridingsKey] = map[string]any{}
	}

	if _, ok := params[kwasm.OperatorChartOverridingsKey].(map[string]any)["kwasmOperator"]; !ok {
		params[kwasm.OperatorChartOverridingsKey].(map[string]any)["kwasmOperator"] = map[string]any{
			"installerImage": nodeInstallerOCI,
		}
	} else {
		params[kwasm.OperatorChartOverridingsKey].(map[string]any)["kwasmOperator"].(map[string]any)["installerImage"] = nodeInstallerOCI
	}

	return nil
}

func setSpinkubeComponentOverridings(p stack.ComponentOverrides, theThing string) (
	version string,
	url string,
	postInstall string,
	err error,
) {
	releases, err := poller.GetSharedPoller().Get("spinkube", "spin-operator")
	if err != nil {
		return
	}
	url = ""
	postInstall = ""

	_version := getSpinkubeComponentOverridings(p)
	version = apps.GetVersionIfItsNotNilAndLatest(_version, releases[0])

	defaultVals := func() {
		url = fmt.Sprintf("https://github.com/spinkube/spin-operator/releases/download/%s/%s", version, theThing)
		postInstall = "https://www.spinkube.dev/docs/topics/"
	}

	defaultVals()
	return
}

func SpinkubeOperatorCrdComponent(params stack.ComponentOverrides) (stack.Component, error) {

	version, url, postInstall, err := setSpinkubeComponentOverridings(params, "spin-operator.crds.yaml")
	if err != nil {
		return stack.Component{}, err
	}

	return spinkubeReturnHelper(version, url, postInstall)
}

func SpinkubeOperatorRuntimeClassComponent(params stack.ComponentOverrides) (stack.Component, error) {

	version, url, postInstall, err := setSpinkubeComponentOverridings(params, "spin-operator.runtime-class.yaml")
	if err != nil {
		return stack.Component{}, err
	}

	return spinkubeReturnHelper(version, url, postInstall)
}

func SpinkubeOperatorShimExecComponent(params stack.ComponentOverrides) (stack.Component, error) {

	version, url, postInstall, err := setSpinkubeComponentOverridings(params, "spin-operator.shim-executor.yaml")
	if err != nil {
		return stack.Component{}, err
	}

	return spinkubeReturnHelper(version, url, postInstall)
}

func spinkubeReturnHelper(version, url, postInstall string) (stack.Component, error) {
	return stack.Component{
		HandlerType: stack.ComponentTypeKubectl,
		Kubectl: &k8s.App{
			Urls:            []string{url},
			Version:         version,
			CreateNamespace: false,
			Metadata:        fmt.Sprintf("KubeSpin (ver: %s) is an open source project that streamlines developing, deploying and operating WebAssembly workloads in Kubernetes - resulting in delivering smaller, more portable applications and incredible compute performance benefits", version),
			PostInstall:     postInstall,
		},
	}, nil
}

const (
	SKU stack.ComponentID = "spinkube"
)

func SpinOperatorComponent(params stack.ComponentOverrides) (stack.Component, error) {

	version, helmOverride := setSpinOperatorComponentOverridings(params)

	if strings.HasPrefix(version, "v") {
		version = strings.TrimPrefix(version, "v")
	}
	return stack.Component{
		HandlerType: stack.ComponentTypeHelm,
		Helm: &helm.App{
			Charts: []helm.ChartOptions{
				{
					Name:            fmt.Sprintf("./spin-operator-%s.tgz", version),
					Version:         version,
					ReleaseName:     "spin-operator",
					Namespace:       "spin-operator",
					CreateNamespace: true,
					Args:            helmOverride,
					ChartRef:        "oci://ghcr.io/spinkube/charts/spin-operator",
				},
			},
		},
	}, nil
}

func getSpinkubeOperatorComponentOverridings(p stack.ComponentOverrides) (version *string, helmOperatorChartOverridings map[string]interface{}) {
	helmOperatorChartOverridings = nil // By default, it is nil

	if p == nil {
		return nil, nil
	}

	for k, v := range p {
		switch k {
		case "version":
			if v, ok := v.(string); ok {
				version = utilities.Ptr(v)
			}
		case "helmOperatorChartOverridings":
			if v, ok := v.(map[string]interface{}); ok {
				helmOperatorChartOverridings = v
			}
		}
	}
	return
}

func setSpinOperatorComponentOverridings(p stack.ComponentOverrides) (
	version string,
	helmOperatorChartOverridings map[string]any,
) {

	releases, err := poller.GetSharedPoller().Get("spinkube", "spin-operator")
	if err != nil {
		return
	}

	helmOperatorChartOverridings = map[string]any{}

	_version, _helmOperatorChartOverridings := getSpinkubeOperatorComponentOverridings(p)

	version = apps.GetVersionIfItsNotNilAndLatest(_version, releases[0])

	if _helmOperatorChartOverridings != nil {
		helmOperatorChartOverridings = _helmOperatorChartOverridings
	}

	return
}
