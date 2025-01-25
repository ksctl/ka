package lite

import (
	"github.com/ksctl/ka/internal/apps/kubeprometheus"
	"github.com/ksctl/ksctl/pkg/apps/stack"
)

const (
	SKU stack.ID = "monitoring-lite"
)

func MonitoringLite(params stack.ApplicationParams) (stack.ApplicationStack, error) {
	stk := stack.ApplicationStack{
		Components: map[stack.ComponentID]stack.Component{
			kubeprometheus.SKU: kubeprometheus.KubePrometheusStandardComponent(
				params.ComponentParams[kubeprometheus.SKU],
			),
		},

		StkDepsIdx:  []stack.ComponentID{kubeprometheus.SKU},
		StackNameID: SKU,
		Maintainer:  "dipankar.das@ksctl.com",
	}

	return stk, nil
}
