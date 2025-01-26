package executor

import (
	"context"
	"os"

	"github.com/ksctl/ksctl/v2/pkg/consts"
	ksctlHelm "github.com/ksctl/ksctl/v2/pkg/helm"
	"github.com/ksctl/ksctl/v2/pkg/logger"
)

func HelmDeployHandler(ctx context.Context, app *ksctlHelm.App) error {
	helmOption := []ksctlHelm.Option{
		ksctlHelm.WithDebug(),
	}
	if v, ok := os.LookupEnv("HELMOCI_CHARTS_DIR"); ok {
		helmOption = append(helmOption, ksctlHelm.WithOCIChartPullDestDir(v))
	}

	obj, err := ksctlHelm.NewClient(
		context.WithValue(ctx, consts.KsctlModuleNameKey, "ksctl.com/helm-client"),
		logger.NewStructuredLogger(-1, os.Stdout),
		helmOption...,
	)
	if err != nil {
		return err
	}

	if err := obj.HelmDeploy(app); err != nil {
		return err
	}

	return nil
}

func HelmUninstallHandler(ctx context.Context, app *ksctlHelm.App) error {
	helmOption := []ksctlHelm.Option{
		ksctlHelm.WithDebug(),
	}
	if v, ok := os.LookupEnv("HELMOCI_CHARTS_DIR"); ok {
		helmOption = append(helmOption, ksctlHelm.WithOCIChartPullDestDir(v))
	}

	obj, err := ksctlHelm.NewClient(
		context.WithValue(ctx, consts.KsctlModuleNameKey, "ksctl.com/helm-client"),
		logger.NewStructuredLogger(-1, os.Stdout),
		helmOption...,
	)
	if err != nil {
		return err
	}

	if err := obj.HelmUninstall(app); err != nil {
		return err
	}
	return nil
}
