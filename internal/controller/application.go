package controller

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/ksctl/ksctl/pkg/logger"
	"sigs.k8s.io/controller-runtime/pkg/client"

	appv1 "github.com/ksctl/ka/api/v1"
	ksctlHelm "github.com/ksctl/ksctl/pkg/helm"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
)

func (r *StackReconciler) GetData(ctx context.Context) (*corev1.ConfigMap, error) {
	cf := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ka-state",
			Namespace: "ksctl-system",
		},
		Data: map[string]string{},
	}

	if err := r.Get(ctx, client.ObjectKey{Namespace: cf.Namespace, Name: cf.Name}, cf); err != nil {
		if errors.IsNotFound(err) {
			cf = &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "ka-state",
					Namespace: "ksctl-system",
				},
				Data: map[string]string{},
			}
			if err := r.Create(ctx, cf); err != nil {
				return nil, err
			}
			return cf, nil
		}
		return nil, err
	}
	if cf.Data == nil {
		cf.Data = map[string]string{}
	}

	return cf, nil
}

func (r *StackReconciler) UpdateData(ctx context.Context, cf *corev1.ConfigMap) error {
	if err := r.Update(ctx, cf); err != nil {
		if errors.IsNotFound(err) {
			return r.Create(ctx, cf)
		}
		return err
	}
	return nil
}

func (r *StackReconciler) InstallApp(ctx context.Context, instance *appv1.Stack) error {

	cf, err := r.GetData(ctx)
	if err != nil {
		return fmt.Errorf("failed to get/create config map: %w", err)
	}

	if _, installed := cf.Data["example"]; installed {
		return nil
	}

	obj, err := ksctlHelm.NewInClusterHelmClient(ctx, logger.NewStructuredLogger(-1, os.Stdout))
	if err != nil {
		return err
	}

	if err := obj.HelmDeploy(&ksctlHelm.App{
		RepoName: "examples",
		RepoUrl:  "https://helm.github.io/examples",
		Charts: []ksctlHelm.ChartOptions{
			{
				ReleaseName: "ahoy",
				Name:        "examples/hello-world",
			},
		},
	}); err != nil {
		return err
	}

	return r.updateAddonStatus(ctx, cf, "example", false)
}

func (r *StackReconciler) UninstallApp(ctx context.Context, instance *appv1.Stack) error {
	cf, err := r.GetData(ctx)
	if err != nil {
		return fmt.Errorf("failed to get/create config map: %w", err)
	}

	if _, installed := cf.Data["example"]; !installed {
		return nil
	}

	obj, err := ksctlHelm.NewInClusterHelmClient(ctx, logger.NewStructuredLogger(-1, os.Stdout))
	if err != nil {
		return err
	}

	if err := obj.HelmUninstall(&ksctlHelm.App{
		RepoName: "examples",
		RepoUrl:  "https://helm.github.io/examples",
		Charts: []ksctlHelm.ChartOptions{
			{
				ReleaseName: "ahoy",
				Name:        "examples/hello-world",
			},
		},
	}); err != nil {
		return err
	}

	return r.updateAddonStatus(ctx, cf, "example", true)
}

func (r *StackReconciler) updateAddonStatus(ctx context.Context, cf *corev1.ConfigMap, appName string, isDelete bool) error {
	return retry.OnError(retry.DefaultRetry, errors.IsConflict, func() error {
		if isDelete {
			delete(cf.Data, appName)
		} else {
			if cf.Data == nil {
				cf.Data = map[string]string{}
			}
			cf.Data[appName] = fmt.Sprintf("installed@%s", time.Now().Format(time.RFC3339))
		}
		return r.Update(ctx, cf)
	})
}
