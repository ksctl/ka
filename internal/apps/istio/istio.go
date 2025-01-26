package istio

import (
	"strings"

	"github.com/ksctl/ka/internal/apps"
	"github.com/ksctl/ksctl/v2/pkg/apps/stack"
	"github.com/ksctl/ksctl/v2/pkg/helm"

	"github.com/ksctl/ksctl/v2/pkg/poller"
	"github.com/ksctl/ksctl/v2/pkg/utilities"
)

func getIstioComponentOverridings(p stack.ComponentOverrides) (version *string, helmBaseChartOverridings map[string]interface{}, helmIstiodChartOverridings map[string]interface{}) {
	helmBaseChartOverridings = nil // By default, it is nil
	helmIstiodChartOverridings = nil

	if p == nil {
		return nil, nil, nil
	}

	for k, v := range p {
		switch k {
		case "version":
			if v, ok := v.(string); ok {
				version = utilities.Ptr(v)
			}
		case "helmBaseChartOverridings":
			if v, ok := v.(map[string]interface{}); ok {
				helmBaseChartOverridings = v
			}
		case "helmIstiodChartOverridings":
			if v, ok := v.(map[string]interface{}); ok {
				helmIstiodChartOverridings = v
			}
		}
	}
	return
}

func setIsitoComponentOverridings(p stack.ComponentOverrides) (
	version string,
	helmBaseChartOverridings map[string]any,
	helmIstiodChartOverridings map[string]any,
	err error,
) {
	releases, err := poller.GetSharedPoller().Get("istio", "istio")
	if err != nil {
		return "", nil, nil, err
	}

	_version, _helmBaseChartOverridings, _helmIstiodChartOverridings := getIstioComponentOverridings(p)

	version = apps.GetVersionIfItsNotNilAndLatest(_version, releases[0])

	if _helmBaseChartOverridings != nil {
		helmBaseChartOverridings = _helmBaseChartOverridings
	} else {
		helmBaseChartOverridings = map[string]any{
			"defaultRevision": "default",
		}
	}

	if _helmIstiodChartOverridings != nil {
		helmIstiodChartOverridings = _helmIstiodChartOverridings
	} else {
		helmIstiodChartOverridings = nil
	}
	return version, helmBaseChartOverridings, helmIstiodChartOverridings, nil
}

const (
	SKU stack.ComponentID = "istio"
)

func IstioStandardComponent(params stack.ComponentOverrides) (stack.Component, error) {

	version, helmBaseChartOverridings, helmIstiodChartOverridings, err := setIsitoComponentOverridings(params)
	if err != nil {
		return stack.Component{}, err
	}

	version = strings.TrimPrefix(version, "v")

	return stack.Component{
		Helm: &helm.App{
			RepoUrl:  "https://istio-release.storage.googleapis.com/charts",
			RepoName: "istio",
			Charts: []helm.ChartOptions{
				{
					Name:            "istio/base",
					Version:         version,
					ReleaseName:     "istio-base",
					Namespace:       "istio-system",
					CreateNamespace: true,
					Args:            helmBaseChartOverridings,
				},
				{
					Name:            "istio/istiod",
					Version:         version,
					ReleaseName:     "istiod",
					Namespace:       "istio-system",
					CreateNamespace: false,
					Args:            helmIstiodChartOverridings,
				},
			},
		},
		HandlerType: stack.ComponentTypeHelm,
	}, nil
}
