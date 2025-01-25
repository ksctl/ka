package executor

import (
	"context"
	"os"

	"github.com/ksctl/ksctl/pkg/consts"
	"github.com/ksctl/ksctl/pkg/k8s"
	"github.com/ksctl/ksctl/pkg/logger"
	"k8s.io/client-go/rest"
)

func K8sDeployHandler(
	ctx context.Context,
	c *rest.Config,
	app *k8s.App,
) error {
	obj, err := k8s.NewK8sClient(context.WithValue(ctx, consts.KsctlModuleNameKey, "ksctl.com/k8s-client"), logger.NewStructuredLogger(-1, os.Stdout), c)
	if err != nil {
		return err
	}

	return obj.KubectlApply(app)
}

func K8sUninstallHandler(
	ctx context.Context,
	c *rest.Config,
	app *k8s.App,
) error {
	obj, err := k8s.NewK8sClient(context.WithValue(ctx, consts.KsctlModuleNameKey, "ksctl.com/k8s-client"), logger.NewStructuredLogger(-1, os.Stdout), c)
	if err != nil {
		return err
	}

	return obj.KubectlDelete(app)
}
