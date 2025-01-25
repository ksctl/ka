package spinkube

import (
	"github.com/ksctl/ka/internal/apps/certmanager"
	"github.com/ksctl/ka/internal/apps/kwasm"
	"github.com/ksctl/ka/internal/apps/spinkube"
	"github.com/ksctl/ksctl/v2/pkg/apps/stack"
)

const (
	SKU stack.ID = "wasm/spinkube-standard"
)

func SpinkubeStandard(params stack.ApplicationParams) (stack.ApplicationStack, error) {

	certManagerComponent, err := certmanager.CertManagerComponent(
		params.ComponentParams[certmanager.SKU],
	)
	if err != nil {
		return stack.ApplicationStack{}, err
	}

	if err := spinkube.GetSpinKubeStackSpecificKwasmOverrides(
		params.ComponentParams[kwasm.OperatorSKU]); err != nil {
		return stack.ApplicationStack{}, err
	}

	kwasmOperatorComponent, err := kwasm.KwasmOperatorComponent(
		params.ComponentParams[kwasm.OperatorSKU],
	)
	if err != nil {
		return stack.ApplicationStack{}, err
	}

	spinKubeCrd, err := spinkube.SpinkubeOperatorCrdComponent(
		params.ComponentParams[spinkube.OperatorCrdSKU],
	)
	if err != nil {
		return stack.ApplicationStack{}, err
	}

	spinKubeRuntime, err := spinkube.SpinkubeOperatorRuntimeClassComponent(
		params.ComponentParams[spinkube.OperatorRuntimeClassSKU],
	)
	if err != nil {
		return stack.ApplicationStack{}, err
	}

	spinKubeShim, err := spinkube.SpinkubeOperatorShimExecComponent(
		params.ComponentParams[spinkube.OperatorShimExecutorSKU],
	)
	if err != nil {
		return stack.ApplicationStack{}, err
	}

	spinOperator, err := spinkube.SpinOperatorComponent(
		params.ComponentParams[spinkube.OperatorSKU],
	)
	if err != nil {
		return stack.ApplicationStack{}, err
	}

	return stack.ApplicationStack{
		Maintainer:  "dipankar.das@ksctl.com",
		StackNameID: SKU,
		Components: map[stack.ComponentID]stack.Component{
			certmanager.SKU:                  certManagerComponent,
			kwasm.OperatorSKU:                kwasmOperatorComponent,
			spinkube.OperatorCrdSKU:          spinKubeCrd,
			spinkube.OperatorRuntimeClassSKU: spinKubeRuntime,
			spinkube.OperatorShimExecutorSKU: spinKubeShim,
			spinkube.OperatorSKU:             spinOperator,
		},
		StkDepsIdx: []stack.ComponentID{
			certmanager.SKU,
			kwasm.OperatorSKU,
			spinkube.OperatorCrdSKU,
			spinkube.OperatorRuntimeClassSKU,
			spinkube.OperatorShimExecutorSKU,
			spinkube.OperatorSKU,
		},
	}, nil
}
