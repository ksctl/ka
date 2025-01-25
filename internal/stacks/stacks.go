package stacks

import (
	"context"

	gitOpsStandard "github.com/ksctl/ka/internal/stacks/gitops"
	meshStandard "github.com/ksctl/ka/internal/stacks/mesh/standard"
	monitoringLite "github.com/ksctl/ka/internal/stacks/monitoring/lite"
	kwasmPlus "github.com/ksctl/ka/internal/stacks/wasm/kwasm"
	spinkubeStandard "github.com/ksctl/ka/internal/stacks/wasm/spinkube"
	"github.com/ksctl/ksctl/pkg/apps/stack"
	ksctlErrors "github.com/ksctl/ksctl/pkg/errors"
	"github.com/ksctl/ksctl/pkg/logger"
)

var stackManifests = map[stack.ID]func(stack.ApplicationParams) (stack.ApplicationStack, error){
	gitOpsStandard.SKU:   gitOpsStandard.GitOps,
	monitoringLite.SKU:   monitoringLite.MonitoringLite,
	meshStandard.SKU:     meshStandard.MeshStandard,
	kwasmPlus.SKU:        kwasmPlus.KwasmPlus,
	spinkubeStandard.SKU: spinkubeStandard.SpinkubeStandard,
}

func Get(ctx context.Context, log logger.Logger, stkID string) (func(stack.ApplicationParams) (stack.ApplicationStack, error), error) {
	fn, ok := stackManifests[stack.ID(stkID)]
	if !ok {
		return nil, ksctlErrors.WrapError(
			ksctlErrors.ErrFailedKsctlComponent,
			log.NewError(ctx, "appStack not found", "stkId", stkID),
		)
	}
	return fn, nil
}

func GetComponentVersionOverriding(component stack.Component) string {
	if component.HandlerType == stack.ComponentTypeKubectl {
		return component.Kubectl.Version
	}
	return component.Helm.Charts[0].Version
}
