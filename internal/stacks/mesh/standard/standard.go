package standard

import (
	"github.com/ksctl/ka/internal/apps/istio"
	"github.com/ksctl/ksctl/pkg/apps/stack"
)

const (
	SKU stack.ID = "mesh-standard"
)

func MeshStandard(params stack.ApplicationParams) (stack.ApplicationStack, error) {

	v, err := istio.IstioStandardComponent(
		params.ComponentParams[istio.SKU],
	)
	if err != nil {
		return stack.ApplicationStack{}, err
	}

	return stack.ApplicationStack{
		Components: map[stack.ComponentID]stack.Component{
			istio.SKU: v,
		},

		StkDepsIdx: []stack.ComponentID{
			istio.SKU,
		},
		Maintainer:  "dipankar.das@ksctl.com",
		StackNameID: SKU,
	}, nil
}
