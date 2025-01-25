package kwasm

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

func TestSetKwasmOperatorComponentOverridings_DefaultValues(t *testing.T) {
	params := stack.ComponentOverrides{}
	version, overridings, err := setKwasmOperatorComponentOverridings(params)

	assert.NoError(t, err)
	assert.Equal(t, "latest", version)
	assert.Nil(t, overridings)
}

func TestSetKwasmOperatorComponentOverridings_WithOverrides(t *testing.T) {
	params := stack.ComponentOverrides{
		"version": "v1.2.3",
		"kwasmOperatorChartOverridings": map[string]any{
			"someKey": "someValue",
		},
	}
	version, overridings, err := setKwasmOperatorComponentOverridings(params)

	assert.NoError(t, err)
	assert.Equal(t, "v1.2.3", version)
	assert.NotNil(t, overridings)
}

func TestKwasmWasmedgeComponent(t *testing.T) {
	params := stack.ComponentOverrides{}
	component, err := KwasmComponent(params)

	assert.NoError(t, err)
	assert.Equal(t, stack.ComponentTypeKubectl, component.HandlerType)
	assert.NotNil(t, component.Kubectl)
	assert.Equal(t, "latest", component.Kubectl.Version)
	assert.Equal(t, []string{"https://raw.githubusercontent.com/ksctl/components/main/wasm/kwasm/runtimeclass.yml"}, component.Kubectl.Urls)
}
