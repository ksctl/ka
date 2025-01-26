package argocd

import (
	"fmt"

	"github.com/ksctl/ka/internal/apps"
	"github.com/ksctl/ksctl/v2/pkg/apps/stack"
	"github.com/ksctl/ksctl/v2/pkg/k8s"
	"github.com/ksctl/ksctl/v2/pkg/utilities"
)

func getArgocdComponentOverridings(p stack.ComponentOverrides) (
	version *string,
	noUI *bool,
	namespaceInstall *bool,
	namespace *string,
) {
	if p == nil {
		return nil, nil, nil, nil
	}
	for k, v := range p {
		switch k {
		case "version":
			if v, ok := v.(string); ok {
				version = utilities.Ptr(v)
			}
		case "noUI":
			if v, ok := v.(bool); ok {
				noUI = utilities.Ptr(v)
			}
		case "namespaceInstall":
			if v, ok := v.(bool); ok {
				namespaceInstall = utilities.Ptr(v)
			}
		case "namespace":
			if v, ok := v.(string); ok {
				namespace = utilities.Ptr(v)
			}
		}
	}
	return
}

func setArgocdComponentOverridings(p stack.ComponentOverrides) (
	version string,
	url []string,
	postInstall string,
	namespace string,
) {
	url = nil
	postInstall = ""
	namespace = "argocd"

	_version, _noUI, _namespaceInstall, _namespace := getArgocdComponentOverridings(p)
	if _namespace != nil {
		if *_namespace != "argocd" {
			namespace = *_namespace
		}
	}

	version = apps.GetVersionIfItsNotNilAndLatest(_version, "stable")

	generateManifestUrl := func(ver string, path string) string {
		return fmt.Sprintf("https://raw.githubusercontent.com/argoproj/argo-cd/%s/%s", ver, path)
	}

	defaultVals := func() {
		url = []string{
			generateManifestUrl(version, "manifests/install.yaml"),
		}
		postInstall = `
Commands to execute to access Argocd
$ kubectl get secret -n argocd argocd-initial-admin-secret -o json | jq -r '.data.password' | base64 -d
$ kubectl port-forward svc/argocd-server -n argocd 8080:443
and login to http://localhost:8080 with user admin and password from above
`
	}

	if _noUI != nil {
		if !*_noUI {
			defaultVals()
		} else {
			url = []string{
				generateManifestUrl(version, "manifests/core-install.yaml"),
			}
			postInstall = fmt.Sprintf(`
https://argo-cd.readthedocs.io/en/%s/operator-manual/core/
`, version)
		}
	} else if _namespaceInstall != nil {
		if *_namespaceInstall {
			url = []string{
				generateManifestUrl(version, "manifests/crds/application-crd.yaml"),
				generateManifestUrl(version, "manifests/crds/appproject-crd.yaml"),
				generateManifestUrl(version, "manifests/crds/applicationset-crd.yaml"),
				generateManifestUrl(version, "manifests/namespace-install.yaml"),
			}
			postInstall = fmt.Sprintf(`
https://argo-cd.readthedocs.io/en/%s/operator-manual/installation/#non-high-availability
`, version)
		} else {
			defaultVals()
		}
	} else {
		defaultVals()
	}

	return version, url, postInstall, namespace
}

const (
	SKU stack.ComponentID = "argocd"
)

func ArgoCDStandardComponent(params stack.ComponentOverrides) stack.Component {
	version, url, postInstall, ns := setArgocdComponentOverridings(params)

	return stack.Component{
		Kubectl: &k8s.App{
			Namespace:       ns,
			CreateNamespace: true,
			Urls:            url,
			Version:         version,
			Metadata:        fmt.Sprintf("Argo CD (Ver: %s) is a declarative, GitOps continuous delivery tool for Kubernetes.", version),
			PostInstall:     postInstall,
		},
		HandlerType: stack.ComponentTypeKubectl,
	}
}
