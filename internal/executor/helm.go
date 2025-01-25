package executor

import (
	"context"
	"os"

	"github.com/ksctl/ksctl/pkg/consts"
	ksctlHelm "github.com/ksctl/ksctl/pkg/helm"
	"github.com/ksctl/ksctl/pkg/logger"
)

func HelmDeployHandler(ctx context.Context, app *ksctlHelm.App) error {

	obj, err := ksctlHelm.NewInClusterHelmClient(context.WithValue(ctx, consts.KsctlModuleNameKey, "ksctl.com/helm-client"), logger.NewStructuredLogger(-1, os.Stdout))
	if err != nil {
		return err
	}

	if err := obj.HelmDeploy(app); err != nil {
		return err
	}
	return nil
}

func HelmUninstallHandler(ctx context.Context, app *ksctlHelm.App) error {

	obj, err := ksctlHelm.NewInClusterHelmClient(context.WithValue(ctx, consts.KsctlModuleNameKey, "ksctl.com/helm-client"), logger.NewStructuredLogger(-1, os.Stdout))
	if err != nil {
		return err
	}

	if err := obj.HelmUninstall(app); err != nil {
		return err
	}
	return nil
}
