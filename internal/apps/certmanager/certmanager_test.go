package certmanager

import (
	"sort"
	"testing"

	"github.com/ksctl/ksctl/v2/pkg/apps/stack"
	"github.com/ksctl/ksctl/v2/pkg/poller"

	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	poller.InitSharedGithubReleaseFakePoller(func(org, repo string) ([]string, error) {
		vers := []string{"v0.0.1"}

		switch org + " " + repo {
		case "cert-manager cert-manager":
			vers = append(vers, "v1.15.3")
		}

		sort.Slice(vers, func(i, j int) bool {
			return vers[i] > vers[j]
		})

		return vers, nil
	})
	m.Run()
}

func TestCertManagerComponentWithNilParams(t *testing.T) {
	params := stack.ComponentOverrides(nil)
	component, err := CertManagerComponent(params)
	assert.NoError(t, err)
	assert.Equal(t, "1.15.3", component.Helm.Charts[0].Version)
	assert.Equal(t, "cert-manager", component.Helm.Charts[0].ReleaseName)
	assert.Equal(t, "cert-manager", component.Helm.Charts[0].Namespace)
	assert.Equal(t, "https://charts.jetstack.io", component.Helm.RepoUrl)
	assert.Equal(t, "jetstack", component.Helm.RepoName)

	if v, ok := component.Helm.Charts[0].Args["crds"]; !ok {
		t.Fatal("missing crds")
	} else {
		if v, ok := v.(map[string]any)["enabled"]; !ok {
			t.Fatal("missing enabled")
		} else {
			assert.Equal(t, true, v)
		}
	}
}

func TestCertManagerComponentWithVersionOverride(t *testing.T) {
	params := stack.ComponentOverrides{
		"version": "v1.0.0",
	}
	component, err := CertManagerComponent(params)
	assert.NoError(t, err)
	assert.Equal(t, "1.0.0", component.Helm.Charts[0].Version)
}

func TestCertManagerComponentWithGatewayApiEnable(t *testing.T) {
	params := stack.ComponentOverrides{
		"gatewayapiEnable": true,
	}
	component, err := CertManagerComponent(params)
	assert.NoError(t, err)
	assert.Equal(t, "1.15.3", component.Helm.Charts[0].Version)
	assert.Contains(t, component.Helm.Charts[0].Args["extraArgs"], "--enable-gateway-api")
}

func TestCertManagerComponentWithCertManagerChartOverridings(t *testing.T) {
	params := stack.ComponentOverrides{
		"certmanagerChartOverridings": map[string]any{
			"someKey": "someValue",
		},
	}
	component, err := CertManagerComponent(params)
	assert.NoError(t, err)
	assert.Equal(t, "1.15.3", component.Helm.Charts[0].Version)
	assert.Equal(t, "someValue", component.Helm.Charts[0].Args["someKey"])

	if v, ok := component.Helm.Charts[0].Args["crds"]; !ok {
		t.Fatal("missing crds")
	} else {
		if v, ok := v.(map[string]any)["enabled"]; !ok {
			t.Fatal("missing enabled")
		} else {
			assert.Equal(t, true, v)
		}
	}
}

func TestCertManagerComponentWithAllOverrides(t *testing.T) {
	params := stack.ComponentOverrides{
		"version":          "v1.0.0",
		"gatewayapiEnable": true,
		"certmanagerChartOverridings": map[string]any{
			"someKey": "someValue",
		},
	}
	component, err := CertManagerComponent(params)
	assert.NoError(t, err)
	assert.Equal(t, "1.0.0", component.Helm.Charts[0].Version)
	assert.Contains(t, component.Helm.Charts[0].Args["extraArgs"], "--enable-gateway-api")
	assert.Equal(t, "someValue", component.Helm.Charts[0].Args["someKey"])
	if v, ok := component.Helm.Charts[0].Args["crds"]; !ok {
		t.Fatal("missing crds")
	} else {
		if v, ok := v.(map[string]any)["enabled"]; !ok {
			t.Fatal("missing enabled")
		} else {
			assert.Equal(t, true, v)
		}
	}
}
