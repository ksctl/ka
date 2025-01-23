package gitops

import (
	"github.com/ksctl/ka/internal/apps/argocd"
	"github.com/ksctl/ka/internal/apps/argorollouts"
	"github.com/ksctl/ksctl/pkg/apps/stack"
)

const (
	SKU stack.ID = "gitops-standard"
)

func GitOps(params stack.ApplicationParams) (stack.ApplicationStack, error) {
	v, err := argorollouts.ArgoRolloutsStandardComponent(
		params.ComponentParams[argorollouts.SKU],
	)
	if err != nil {
		return stack.ApplicationStack{}, err
	}

	stk := stack.ApplicationStack{
		Components: map[stack.ComponentID]stack.Component{
			argocd.SKU: argocd.ArgoCDStandardComponent(
				params.ComponentParams[argocd.SKU],
			),
			argorollouts.SKU: v,
		},

		StkDepsIdx:  []stack.ComponentID{argocd.SKU, argorollouts.SKU},
		StackNameID: SKU,
		Maintainer:  "dipankar.das@ksctl.com",
	}

	return stk, nil
}
