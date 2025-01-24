package executor

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/gookit/goutil/dump"
	"github.com/ksctl/ksctl/pkg/k8s"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/client"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type handler struct {
	ctx context.Context
	c   client.Client
	d   dynamic.Interface
	r   meta.RESTMapper
	s   *runtime.Scheme
}

func K8sDeployHandler(
	ctx context.Context,
	c client.Client,
	d dynamic.Interface,
	r meta.RESTMapper,
	s *runtime.Scheme,
	app *k8s.App,
) error {
	h := handler{ctx, c, d, r, s}

	return h.manifestsInstall(app)
}

func K8sUninstallHandler(
	ctx context.Context,
	c client.Client,
	d dynamic.Interface,
	r meta.RESTMapper,
	s *runtime.Scheme,
	app *k8s.App,
) error {
	h := handler{ctx, c, d, r, s}

	return h.manifestsUninstall(app)
}

func (h *handler) manifestsInstall(app *k8s.App) error {
	ns := func() *string {
		if app.CreateNamespace {
			return &app.Namespace
		}
		return nil
	}()

	if app.CreateNamespace {
		if err := h.CreateNamespaceIfNotExists(h.ctx, *ns); err != nil {
			return fmt.Errorf("failed to create namespace for component %s: %w", *ns, err)
		}
	}

	for _, url := range app.Urls {
		if err := h.downloadManifest(h.ctx, url, ns, h.applyResource); err != nil {
			return err
		}
	}

	return nil
}

func (h *handler) manifestsUninstall(app *k8s.App) error {
	ns := func() *string {
		if app.CreateNamespace {
			return &app.Namespace
		}
		return nil
	}()

	for i := len(app.Urls) - 1; i >= 0; i-- {
		url := app.Urls[i]
		if err := h.downloadManifest(h.ctx, url, ns, h.applyResource); err != nil {
			return err
		}
	}

	if app.CreateNamespace {
		if err := h.DeleteNamespaceIfExists(h.ctx, *ns); err != nil {
			return fmt.Errorf("failed to delete namespace for component %s: %w", *ns, err)
		}
	}

	return nil
}

func (h *handler) downloadManifest(
	ctx context.Context,
	url string,
	namespace *string,
	operator func(ctx context.Context, obj *unstructured.Unstructured) error,
) error {

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download manifest: %w", err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download manifest, status: %d", resp.StatusCode)
	}

	decoder := yaml.NewYAMLOrJSONDecoder(resp.Body, 4096)
	for {
		var rawObj map[string]interface{}
		if err := decoder.Decode(&rawObj); err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("failed to decode manifest: %w", err)
		}

		if len(rawObj) == 0 {
			continue // Skip empty documents
		}

		obj := &unstructured.Unstructured{Object: rawObj}

		// Set namespace for namespaced resources if not specified
		if obj.GetNamespace() == "" && namespace != nil {
			obj.SetNamespace(*namespace)
		}

		// Validate required fields
		if obj.GetAPIVersion() == "" || obj.GetKind() == "" {
			return fmt.Errorf("manifest missing apiVersion or kind")
		}

		if err := operator(ctx, obj); err != nil {
			return fmt.Errorf("failed to apply resource %s/%s: %w",
				obj.GetNamespace(), obj.GetName(), err)
		}
	}

	return nil
}

func (r *handler) CreateNamespaceIfNotExists(ctx context.Context, namespace string) error {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}
	if err := r.c.Get(ctx, client.ObjectKey{Name: namespace}, ns); err != nil {
		if errors.IsNotFound(err) {
			return r.c.Create(ctx, ns)
		}
		return err
	}
	return nil
}

func (r *handler) DeleteNamespaceIfExists(ctx context.Context, namespace string) error {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}
	if err := r.c.Get(ctx, client.ObjectKey{Name: namespace}, ns); err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return err
	}
	return r.c.Delete(ctx, ns)
}

func (r *handler) deleteResource(ctx context.Context, obj *unstructured.Unstructured) error {
	// Get the GVK for the resource
	gvk := obj.GroupVersionKind()

	// Get the corresponding REST mapping
	mapping, err := r.r.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return fmt.Errorf("failed to get REST mapping: %w", err)
	}

	// Create dynamic resource interface
	var dr dynamic.ResourceInterface
	if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
		// Namespaced resources
		dr = r.d.Resource(mapping.Resource).Namespace(obj.GetNamespace())
	} else {
		// Cluster-scoped resources
		dr = r.d.Resource(mapping.Resource)
	}

	opts := metav1.DeleteOptions{}

	err = dr.Delete(ctx, obj.GetName(), opts)
	if err != nil {
		fmt.Println("########### Failed to delete resource", obj.GetNamespace(), obj.GetName())
		dump.Println(obj)
		return fmt.Errorf("failed to delete resource: %w", err)
	}

	fmt.Println("########### Delete resource", obj.GetNamespace(), obj.GetName())

	return nil
}

func (r *handler) applyResource(ctx context.Context, obj *unstructured.Unstructured) error {
	// Get the GVK for the resource
	gvk := obj.GroupVersionKind()

	// Get the corresponding REST mapping
	mapping, err := r.r.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return fmt.Errorf("failed to get REST mapping: %w", err)
	}

	// Create dynamic resource interface
	var dr dynamic.ResourceInterface
	if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
		// Namespaced resources
		dr = r.d.Resource(mapping.Resource).Namespace(obj.GetNamespace())
	} else {
		// Cluster-scoped resources
		dr = r.d.Resource(mapping.Resource)
	}

	// Apply the resource using server-side apply
	opts := metav1.ApplyOptions{
		FieldManager: "cluster-addon-controller",
		Force:        true,
	}

	_, err = dr.Apply(ctx, obj.GetName(), obj, opts)
	if err != nil {
		fmt.Println("########### Failed to apply resource", obj.GetNamespace(), obj.GetName())
		dump.Println(obj)
		return fmt.Errorf("failed to apply resource: %w", err)
	}

	fmt.Println("########### Applied resource", obj.GetNamespace(), obj.GetName())

	return nil
}
