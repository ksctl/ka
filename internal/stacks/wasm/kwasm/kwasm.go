package kwasm

import (
	"github.com/ksctl/ka/internal/apps/kwasm"
	"github.com/ksctl/ksctl/v2/pkg/apps/stack"
)

const (
	SKU stack.ID = "wasm/kwasm-plus"
)

func KwasmPlus(params stack.ApplicationParams) (stack.ApplicationStack, error) {

	kwasmOperatorComponent, err := kwasm.KwasmOperatorComponent(
		params.ComponentParams[kwasm.OperatorSKU],
	)
	if err != nil {
		return stack.ApplicationStack{}, err
	}

	wasmedgeWasmtimeKwasmComponent, err := kwasm.KwasmComponent(
		params.ComponentParams[kwasm.RuntimeSKU],
	)
	if err != nil {
		return stack.ApplicationStack{}, err
	}

	return stack.ApplicationStack{
		Maintainer:  "dipankar.das@ksctl.com",
		StackNameID: SKU,
		Components: map[stack.ComponentID]stack.Component{
			kwasm.OperatorSKU: kwasmOperatorComponent,
			kwasm.RuntimeSKU:  wasmedgeWasmtimeKwasmComponent,
		},
		StkDepsIdx: []stack.ComponentID{
			kwasm.OperatorSKU,
			kwasm.RuntimeSKU,
		},
	}, nil
}
