package kubeprometheus

import (
	"github.com/ksctl/ka/internal/apps"
	"github.com/ksctl/ksctl/v2/pkg/apps/stack"
	"github.com/ksctl/ksctl/v2/pkg/helm"
	"strings"

	"github.com/ksctl/ksctl/v2/pkg/utilities"
)

func getKubePrometheusComponentOverridings(p stack.ComponentOverrides) (version *string, helmKubePromChartOverridings map[string]interface{}) {
	helmKubePromChartOverridings = nil // By default it is nil

	if p == nil {
		return nil, nil
	}

	for k, v := range p {
		switch k {
		case "version":
			if v, ok := v.(string); ok {
				version = utilities.Ptr(v)
			}
		case "helmKubePromChartOverridings":
			if v, ok := v.(map[string]interface{}); ok {
				helmKubePromChartOverridings = v
			}
		}
	}
	return
}

func setKubePrometheusComponentOverridings(p stack.ComponentOverrides) (
	version string,
	helmKubePromChartOverridings map[string]any,
) {
	helmKubePromChartOverridings = map[string]any{}

	_version, _helmKubePromChartOverridings := getKubePrometheusComponentOverridings(p)
	version = apps.GetVersionIfItsNotNilAndLatest(_version, "latest")

	if _helmKubePromChartOverridings != nil {
		helmKubePromChartOverridings = _helmKubePromChartOverridings
	} else {
		helmKubePromChartOverridings = nil
	}

	return
}

const (
	SKU stack.ComponentID = "kube-prometheus"
)

func KubePrometheusStandardComponent(params stack.ComponentOverrides) stack.Component {

	version, helmKubePromChartOverridings := setKubePrometheusComponentOverridings(params)

	if strings.HasPrefix(version, "v") {
		version = strings.TrimPrefix(version, "v")
	}

	return stack.Component{
		Helm: &helm.App{
			RepoUrl:  "https://prometheus-community.github.io/helm-charts",
			RepoName: "prometheus-community",
			Charts: []helm.ChartOptions{
				{
					Name:            "prometheus-community/kube-prometheus-stack",
					Version:         version,
					ReleaseName:     "kube-prometheus-stack",
					Namespace:       "monitoring",
					CreateNamespace: true,
					Args:            helmKubePromChartOverridings,
				},
			},
		},
		HandlerType: stack.ComponentTypeHelm,
	}
}
