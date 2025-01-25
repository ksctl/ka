package certmanager

import (
	"slices"
	"strings"

	"github.com/ksctl/ka/internal/apps"
	"github.com/ksctl/ksctl/v2/pkg/apps/stack"
	"github.com/ksctl/ksctl/v2/pkg/helm"

	"github.com/ksctl/ksctl/v2/pkg/poller"

	"github.com/ksctl/ksctl/v2/pkg/utilities"
)

func getCertManagerComponentOverridings(p stack.ComponentOverrides) (
	version *string,
	gateway_apiEnable *bool,
	certmanagerChartOverridings map[string]any,
) {
	if p == nil {
		return nil, nil, nil
	}
	certmanagerChartOverridings = nil

	for k, v := range p {
		switch k {
		case "version":
			if v, ok := v.(string); ok {
				version = utilities.Ptr(v)
			}
		case "certmanagerChartOverridings":
			if v, ok := v.(map[string]any); ok {
				certmanagerChartOverridings = v
			}
		case "gatewayapiEnable":
			if v, ok := v.(bool); ok {
				gateway_apiEnable = utilities.Ptr(v)
			}
		}
	}
	return
}

func setCertManagerComponentOverridings(params stack.ComponentOverrides) (
	version string,
	overridings map[string]any,
	err error,
) {

	releases, err := poller.GetSharedPoller().Get("cert-manager", "cert-manager")
	if err != nil {
		return "", nil, err
	}

	overridings = map[string]any{
		"crds": map[string]any{
			"enabled": true,
		},
	}

	_version, _gateway_apiEnable, _certmanagerChartOverridings := getCertManagerComponentOverridings(params)

	version = apps.GetVersionIfItsNotNilAndLatest(_version, releases[0])

	if _certmanagerChartOverridings != nil {
		utilities.CopySrcToDestPreservingDestVals(overridings, _certmanagerChartOverridings)
	}

	if _gateway_apiEnable != nil { // TODO: need to see later on how
		if *_gateway_apiEnable {
			if v, ok := overridings["extraArgs"]; ok {
				if v, ok := v.([]string); ok {
					if ok := slices.Contains[[]string, string](v, "--enable-gateway-api"); !ok {
						overridings["extraArgs"] = append(v, "--enable-gateway-api")
					}
				}
			} else {
				overridings["extraArgs"] = []string{"--enable-gateway-api"}
			}
		}
	}

	return
}

const (
	SKU stack.ComponentID = "cert-manager"
)

func CertManagerComponent(params stack.ComponentOverrides) (stack.Component, error) {
	version, overridings, err := setCertManagerComponentOverridings(params)
	if err != nil {
		return stack.Component{}, err
	}

	if strings.HasPrefix(version, "v") {
		version = strings.TrimPrefix(version, "v")
	}

	return stack.Component{
		HandlerType: stack.ComponentTypeHelm,
		Helm: &helm.App{
			RepoUrl:  "https://charts.jetstack.io",
			RepoName: "jetstack",
			Charts: []helm.ChartOptions{
				{
					Name:            "jetstack/cert-manager",
					Version:         version,
					ReleaseName:     "cert-manager",
					CreateNamespace: true,
					Namespace:       "cert-manager",
					Args:            overridings,
				},
			},
		},
	}, nil
}
