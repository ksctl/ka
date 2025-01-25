package controller

import (
	"context"
	"encoding/json"
	"os"
	"slices"

	appv1 "github.com/ksctl/ka/api/v1"
	"github.com/ksctl/ka/internal/executor"
	"github.com/ksctl/ka/internal/stacks"
	"github.com/ksctl/ka/internal/stacks/wasm"
	"github.com/ksctl/ksctl/v2/pkg/apps/stack"
	ksctlErrors "github.com/ksctl/ksctl/v2/pkg/errors"
	"github.com/ksctl/ksctl/v2/pkg/logger"

	"sigs.k8s.io/controller-runtime/pkg/log"
)

func GetStackManifest(kl logger.Logger, app *appv1.Stack) (stack.ApplicationStack, error) {

	appStk, err := stacks.Get(context.Background(), kl, app.Spec.StackName)
	if err != nil {
		return stack.ApplicationStack{}, err
	}
	convertedOverriding := make(map[stack.ComponentID]stack.ComponentOverrides)

	var _overrides map[string]map[string]any = nil
	if app.Spec.Overrides != nil {
		_overrides = make(map[string]map[string]any)
		if err := json.Unmarshal(app.Spec.Overrides.Raw, &_overrides); err != nil {
			return stack.ApplicationStack{}, err
		}

		for k, v := range _overrides {
			convertedOverriding[stack.ComponentID(k)] = stack.ComponentOverrides(v)
		}
	}
	return appStk(stack.ApplicationParams{ComponentParams: convertedOverriding})
}

func (r *StackReconciler) Remove(ctx context.Context, app *appv1.Stack) error {
	l := log.FromContext(ctx)
	kl := logger.NewStructuredLogger(-1, os.Stdout)

	if !r.WasStackInstalled(app.Spec.StackName) {
		l.Info("Already uninstalled", "stack", app.Spec.StackName)
		return nil
	}

	manifest, err := GetStackManifest(kl, app)
	if err != nil {
		return err
	}

	defer func() {
		if err := r.Save(ctx); err != nil {
			l.Error(err, "Failed to save state")
		}
	}()

	for i := len(manifest.StkDepsIdx) - 1; i >= 0; i-- {
		componentId := manifest.StkDepsIdx[i]

		if !r.WasComponentInstalled(app.Spec.StackName, string(componentId)) {
			l.Info("Already uninstalled", "component", componentId, "stack", app.Spec.StackName)
			continue
		}

		if slices.Contains(app.Spec.DisableComponents, string(componentId)) {
			l.Info("Component disabled", "component", componentId, "stack", app.Spec.StackName)
			continue
		}

		if v, ok := manifest.Components[componentId]; !ok {
			return ksctlErrors.WrapError(
				ksctlErrors.ErrFailedKsctlComponent,
				kl.NewError(context.Background(), "component not found", "componentId", componentId),
			)
		} else {
			ver := stacks.GetComponentVersionOverriding(v)
			l.Info("Component", "Name", componentId, "Version", ver)
			if v.HandlerType == stack.ComponentTypeKubectl {
				if k8sErr := executor.K8sUninstallHandler(
					ctx,
					r.RestConfig,
					v.Kubectl,
				); k8sErr != nil {
					return k8sErr
				}
			} else {
				if helmErr := executor.HelmUninstallHandler(
					ctx,
					v.Helm,
				); helmErr != nil {
					return helmErr
				}
			}
			delete(r.state.Stacks[app.Spec.StackName].Components, string(componentId))
		}
	}
	if wasm.ShouldPerformAdditionalProcessing(stack.ID(app.Spec.StackName)) {
		if err := wasm.AfterRemoval(ctx, r.Client); err != nil {
			l.Error(err, "Failed to perform additional processing", "purpose", "wasm/node-annotate")
			return err
		}
	}
	delete(r.state.Stacks, app.Spec.StackName)
	l.Info("Successfully uninstalled", "stack", app.Spec.StackName)
	return nil
}

func (r *StackReconciler) Add(ctx context.Context, app *appv1.Stack) error {
	l := log.FromContext(ctx)
	kl := logger.NewStructuredLogger(-1, os.Stdout)

	manifest, err := GetStackManifest(kl, app)
	if err != nil {
		return err
	}

	var appState AppState
	if r.WasStackInstalled(app.Spec.StackName) {
		l.Info("Already installed checking for components", "stack", app.Spec.StackName)
		appState = r.state.Stacks[app.Spec.StackName]
	} else {
		appState = AppState{
			Components: map[string]ComponentState{},
		}
	}

	defer func() {
		r.state.Stacks[app.Spec.StackName] = appState
		if err := r.Save(ctx); err != nil {
			l.Error(err, "Failed to save state")
		}
	}()

	for _, componentId := range manifest.StkDepsIdx {
		if r.WasComponentInstalled(app.Spec.StackName, string(componentId)) {
			l.Info("Already installed", "component", componentId, "stack", app.Spec.StackName)
			continue
		}
		if slices.Contains(app.Spec.DisableComponents, string(componentId)) {
			l.Info("Component disabled", "component", componentId, "stack", app.Spec.StackName)
			continue
		}
		if v, ok := manifest.Components[componentId]; !ok {
			return ksctlErrors.WrapError(
				ksctlErrors.ErrFailedKsctlComponent,
				kl.NewError(context.Background(), "component not found", "componentId", componentId),
			)
		} else {
			ver := stacks.GetComponentVersionOverriding(v)
			l.Info("Component", "Name", componentId, "Version", ver)
			if v.HandlerType == stack.ComponentTypeKubectl {
				if k8sErr := executor.K8sDeployHandler(
					ctx,
					r.RestConfig,
					v.Kubectl,
				); k8sErr != nil {
					return k8sErr
				}
			} else {
				if helmErr := executor.HelmDeployHandler(
					ctx,
					v.Helm,
				); helmErr != nil {
					return helmErr
				}
			}
			appState.Components[string(componentId)] = ComponentState{
				Ver: ver,
			}
		}
	}
	if wasm.ShouldPerformAdditionalProcessing(stack.ID(app.Spec.StackName)) {
		if err := wasm.AfterInstall(ctx, r.Client); err != nil {
			l.Error(err, "Failed to perform additional processing", "purpose", "wasm/node-annotate")
			return err
		}
	}

	l.Info("Successfully installed", "stack", app.Spec.StackName)
	return nil
}
