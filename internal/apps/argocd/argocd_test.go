package argocd

import (
	"github.com/ksctl/ksctl/v2/pkg/apps/stack"
	"github.com/ksctl/ksctl/v2/pkg/poller"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	poller.InitSharedGithubReleaseFakePoller(func(org, repo string) ([]string, error) {
		vers := []string{"v0.0.1"}

		return vers, nil
	})
	m.Run()
}

func TestArgocdComponentOverridingsWithNilParams(t *testing.T) {
	version, url, postInstall, ns := setArgocdComponentOverridings(nil)
	assert.Equal(t, "stable", version)
	assert.Equal(t, "argocd", ns)
	assert.Equal(t, []string{"https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml"}, url)
	assert.Contains(t, postInstall, "Commands to execute to access Argocd")
}

func TestArgocdComponentOverridingsWithEmptyParams(t *testing.T) {
	params := stack.ComponentOverrides{
		"version":   "latest",
		"namespace": "nice",
	}
	version, url, postInstall, ns := setArgocdComponentOverridings(params)
	assert.Equal(t, "nice", ns)
	assert.Equal(t, "stable", version)
	assert.Equal(t, []string{"https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml"}, url)
	assert.Contains(t, postInstall, "Commands to execute to access Argocd")
}

func TestArgocdComponentOverridingsWithVersionOnly(t *testing.T) {
	params := stack.ComponentOverrides{
		"version": "v1.0.0",
	}
	version, url, postInstall, ns := setArgocdComponentOverridings(params)
	assert.Equal(t, "v1.0.0", version)
	assert.Equal(t, "argocd", ns)
	assert.Equal(t, []string{"https://raw.githubusercontent.com/argoproj/argo-cd/v1.0.0/manifests/install.yaml"}, url)
	assert.Contains(t, postInstall, "Commands to execute to access Argocd")
}

func TestArgocdComponentOverridingsWithNoUITrue(t *testing.T) {
	params := stack.ComponentOverrides{
		"noUI": true,
	}
	version, url, postInstall, ns := setArgocdComponentOverridings(params)
	assert.Equal(t, "stable", version)
	assert.Equal(t, "argocd", ns)
	assert.Equal(t, []string{"https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/core-install.yaml"}, url)
	assert.Contains(t, postInstall, "https://argo-cd.readthedocs.io/en/stable/operator-manual/core/")
}

func TestArgocdComponentOverridingsWithNoUIFalse(t *testing.T) {
	params := stack.ComponentOverrides{
		"noUI": false,
	}
	version, url, postInstall, ns := setArgocdComponentOverridings(params)
	assert.Equal(t, "stable", version)
	assert.Equal(t, "argocd", ns)
	assert.Equal(t, []string{"https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml"}, url)
	assert.Contains(t, postInstall, "Commands to execute to access Argocd")
}

func TestArgocdComponentOverridingsWithNamespaceInstallTrue(t *testing.T) {
	params := stack.ComponentOverrides{
		"namespaceInstall": true,
	}
	version, url, postInstall, ns := setArgocdComponentOverridings(params)
	assert.Equal(t, "stable", version)
	assert.Equal(t, "argocd", ns)
	assert.Equal(t,
		[]string{
			"https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/crds/application-crd.yaml",
			"https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/crds/appproject-crd.yaml",
			"https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/crds/applicationset-crd.yaml",
			"https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/namespace-install.yaml",
		}, url)
	assert.Contains(t, postInstall, "https://argo-cd.readthedocs.io/en/stable/operator-manual/installation/#non-high-availability")
}

func TestArgocdComponentOverridingsWithNamespaceInstallFalse(t *testing.T) {
	params := stack.ComponentOverrides{
		"namespaceInstall": false,
	}
	version, url, postInstall, ns := setArgocdComponentOverridings(params)
	assert.Equal(t, "stable", version)
	assert.Equal(t, "argocd", ns)
	assert.Equal(t, []string{"https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml"}, url)
	assert.Contains(t, postInstall, "Commands to execute to access Argocd")
}

func TestArgocdComponentOverridingsWithVersionAndNoUI(t *testing.T) {
	params := stack.ComponentOverrides{
		"version": "v1.0.0",
		"noUI":    true,
	}
	version, url, postInstall, ns := setArgocdComponentOverridings(params)
	assert.Equal(t, "v1.0.0", version)
	assert.Equal(t, "argocd", ns)
	assert.Equal(t, []string{"https://raw.githubusercontent.com/argoproj/argo-cd/v1.0.0/manifests/core-install.yaml"}, url)
	assert.Contains(t, postInstall, "https://argo-cd.readthedocs.io/en/v1.0.0/operator-manual/core/")
}

func TestArgocdComponentOverridingsWithVersionAndNamespaceInstall(t *testing.T) {
	params := stack.ComponentOverrides{
		"version":          "v1.0.0",
		"namespaceInstall": true,
	}
	version, url, postInstall, ns := setArgocdComponentOverridings(params)
	assert.Equal(t, "v1.0.0", version)
	assert.Equal(t, "argocd", ns)
	assert.Equal(t,
		[]string{
			"https://raw.githubusercontent.com/argoproj/argo-cd/v1.0.0/manifests/crds/application-crd.yaml",
			"https://raw.githubusercontent.com/argoproj/argo-cd/v1.0.0/manifests/crds/appproject-crd.yaml",
			"https://raw.githubusercontent.com/argoproj/argo-cd/v1.0.0/manifests/crds/applicationset-crd.yaml",
			"https://raw.githubusercontent.com/argoproj/argo-cd/v1.0.0/manifests/namespace-install.yaml",
		}, url)
	assert.Contains(t, postInstall, "https://argo-cd.readthedocs.io/en/v1.0.0/operator-manual/installation/#non-high-availability")
}
