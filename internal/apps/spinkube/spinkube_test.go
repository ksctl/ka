package spinkube

import (
	"github.com/ksctl/ka/internal/apps/kwasm"
	"github.com/ksctl/ksctl/v2/pkg/apps/stack"
	"github.com/ksctl/ksctl/v2/pkg/poller"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	poller.InitSharedGithubReleaseFakePoller(func(org, repo string) ([]string, error) {
		vers := []string{"v0.0.1"}

		switch org + " " + repo {
		case "spinkube spin-operator":
			vers = append(vers, "v0.2.0")
		case "spinkube containerd-shim-spin":
			vers = append(vers, "v0.15.1")
		}

		sort.Slice(vers, func(i, j int) bool {
			return vers[i] > vers[j]
		})

		return vers, nil
	})
	m.Run()
}

func TestGetSpinKubeStackSpecificKwasmOverrides_DefaultValues(t *testing.T) {
	params := stack.ComponentOverrides{}
	err := GetSpinKubeStackSpecificKwasmOverrides(params)

	assert.NoError(t, err)
	assert.NotNil(t, params[kwasm.OperatorChartOverridingsKey])
	assert.NotNil(t, params[kwasm.OperatorChartOverridingsKey].(map[string]any)["kwasmOperator"])
	assert.Equal(t, "ghcr.io/spinkube/containerd-shim-spin/node-installer:v0.15.1", params[kwasm.OperatorChartOverridingsKey].(map[string]any)["kwasmOperator"].(map[string]any)["installerImage"])
}

func TestGetSpinKubeStackSpecificKwasmOverrides_WithExistingOverrides(t *testing.T) {
	params := stack.ComponentOverrides{
		kwasm.OperatorChartOverridingsKey: map[string]any{
			"kwasmOperator": map[string]any{
				"installerImage": "existing-image",
			},
		},
	}
	err := GetSpinKubeStackSpecificKwasmOverrides(params)

	assert.NoError(t, err)
	assert.Equal(t,
		"ghcr.io/spinkube/containerd-shim-spin/node-installer:v0.15.1",
		params[kwasm.OperatorChartOverridingsKey].(map[string]any)["kwasmOperator"].(map[string]any)["installerImage"])
}

func TestGetSpinKubeStackSpecificKwasmOverrides_NilParams(t *testing.T) {
	err := GetSpinKubeStackSpecificKwasmOverrides(nil)

	assert.NoError(t, err)
}

func TestSpinkubeComponentOverridings_DefaultValues(t *testing.T) {
	params := stack.ComponentOverrides{}
	version, url, postInstall, err := setSpinkubeComponentOverridings(params, "spin-operator.crds.yaml")

	assert.NoError(t, err)
	assert.Equal(t, "v0.2.0", version)
	assert.Equal(t, "https://github.com/spinkube/spin-operator/releases/download/v0.2.0/spin-operator.crds.yaml", url)
	assert.Equal(t, "https://www.spinkube.dev/docs/topics/", postInstall)
}

func TestSpinkubeComponentOverridings_WithOverrides(t *testing.T) {
	params := stack.ComponentOverrides{
		"version": "v1.2.3",
	}
	version, url, postInstall, err := setSpinkubeComponentOverridings(params, "spin-operator.crds.yaml")

	assert.NoError(t, err)
	assert.Equal(t, "v1.2.3", version)
	assert.Equal(t, "https://github.com/spinkube/spin-operator/releases/download/v1.2.3/spin-operator.crds.yaml", url)
	assert.Equal(t, "https://www.spinkube.dev/docs/topics/", postInstall)
}

func TestSpinOperatorComponentOverridings_DefaultValues(t *testing.T) {
	params := stack.ComponentOverrides{}
	version, helmOverride := setSpinOperatorComponentOverridings(params)

	assert.Equal(t, "v0.2.0", version)
	assert.NotNil(t, helmOverride)
}

func TestSpinOperatorComponentOverridings_WithOverrides(t *testing.T) {
	t.Run("WithVersion having v as prefix", func(t *testing.T) {
		params := stack.ComponentOverrides{
			"version": "v1.2.3",
			"helmOperatorChartOverridings": map[string]any{
				"someKey": "someValue",
			},
		}
		version, helmOverride := setSpinOperatorComponentOverridings(params)

		assert.Equal(t, "v1.2.3", version)
		assert.NotNil(t, helmOverride)
		assert.Equal(t, "someValue", helmOverride["someKey"])
	})
	t.Run("WithVersion without having v as prefix", func(t *testing.T) {
		params := stack.ComponentOverrides{
			"version": "1.2.3",
			"helmOperatorChartOverridings": map[string]any{
				"someKey": "someValue",
			},
		}
		version, helmOverride := setSpinOperatorComponentOverridings(params)

		assert.Equal(t, "1.2.3", version)
		assert.NotNil(t, helmOverride)
		assert.Equal(t, "someValue", helmOverride["someKey"])
	})
}
