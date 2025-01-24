package kwasm

import (
	"github.com/ksctl/ka/internal/apps"
	"github.com/ksctl/ksctl/pkg/apps/stack"
	"github.com/ksctl/ksctl/pkg/helm"
	"github.com/ksctl/ksctl/pkg/k8s"
	"strings"

	"github.com/ksctl/ksctl/pkg/utilities"
)

const (
	OperatorChartOverridingsKey string            = "kwasmOperatorChartOverridings"
	SKU                         stack.ComponentID = "kwasm"
)

func getKwasmOperatorComponentOverridings(p stack.ComponentOverrides) (
	version *string,
	kwasmOperatorChartOverridings map[string]any,
) {
	kwasmOperatorChartOverridings = nil // By default it is nil
	if p == nil {
		return nil, nil
	}

	if v, ok := p["version"]; ok {
		if v, ok := v.(string); ok {
			version = utilities.Ptr(v)
		}
	}

	if v, ok := p[OperatorChartOverridingsKey]; ok {
		if v, ok := v.(map[string]any); ok {
			kwasmOperatorChartOverridings = v
		}
	}

	return
}

func setKwasmOperatorComponentOverridings(params stack.ComponentOverrides) (
	version string,
	overridings map[string]any,
	err error,
) {

	_version, _kwasmOperatorChartOverridings := getKwasmOperatorComponentOverridings(params)

	version = apps.GetVersionIfItsNotNilAndLatest(_version, "latest")

	if _kwasmOperatorChartOverridings != nil {
		overridings = utilities.DeepCopyMap(_kwasmOperatorChartOverridings)
	}

	return
}

func KwasmComponent(params stack.ComponentOverrides) (stack.Component, error) {
	return stack.Component{
		HandlerType: stack.ComponentTypeKubectl,
		Kubectl: &k8s.App{
			CreateNamespace: false,
			Version:         "latest",
			Urls:            []string{"https://raw.githubusercontent.com/ksctl/components/main/wasm/kwasm/runtimeclass.yml"},
			Metadata:        "It applies the runtime class for kwasm currently wasmedge and wasmtime",
		},
	}, nil
}

func KwasmOperatorComponent(params stack.ComponentOverrides) (stack.Component, error) {
	version, kwasmOperatorChartOverridings, err := setKwasmOperatorComponentOverridings(params)
	if err != nil {
		return stack.Component{}, err
	}

	if strings.HasPrefix(version, "v") {
		version = strings.TrimPrefix(version, "v")
	}

	return stack.Component{
		Helm: &helm.App{
			RepoName: "kwasm",
			RepoUrl:  "http://kwasm.sh/kwasm-operator/",
			Charts: []helm.ChartOptions{
				{
					Name:            "kwasm/kwasm-operator",
					Version:         version,
					ReleaseName:     "kwasm-operator",
					Namespace:       "kwasm",
					CreateNamespace: true,
					Args:            kwasmOperatorChartOverridings,
				},
			},
		},
	}, nil
}
