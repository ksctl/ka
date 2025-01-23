package kubeprometheus

import (
	"github.com/ksctl/ksctl/pkg/apps/stack"
	"github.com/ksctl/ksctl/pkg/poller"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMain(m *testing.M) {
	poller.InitSharedGithubReleaseFakePoller(func(org, repo string) ([]string, error) {
		vers := []string{"v0.0.1"}

		return vers, nil
	})
	m.Run()
}

func TestKubePrometheusComponentOverridingsWithNilParams(t *testing.T) {
	version, helmKubePromChartOverridings := setKubePrometheusComponentOverridings(nil)
	assert.Equal(t, "latest", version)
	assert.Nil(t, helmKubePromChartOverridings)
}

func TestKubePrometheusComponentOverridingsWithEmptyParams(t *testing.T) {
	params := stack.ComponentOverrides{}
	version, helmKubePromChartOverridings := setKubePrometheusComponentOverridings(params)
	assert.Equal(t, "latest", version)
	assert.Nil(t, helmKubePromChartOverridings)
}

func TestKubePrometheusComponentOverridingsWithVersionOnly(t *testing.T) {
	params := stack.ComponentOverrides{
		"version": "v1.0.0",
	}
	version, helmKubePromChartOverridings := setKubePrometheusComponentOverridings(params)
	assert.Equal(t, "v1.0.0", version)
	assert.Nil(t, helmKubePromChartOverridings)
}

func TestKubePrometheusComponentOverridingsWithHelmKubePromChartOverridingsOnly(t *testing.T) {
	params := stack.ComponentOverrides{
		"helmKubePromChartOverridings": map[string]any{"key": "value"},
	}
	version, helmKubePromChartOverridings := setKubePrometheusComponentOverridings(params)
	assert.Equal(t, "latest", version)
	assert.Equal(t, map[string]any{"key": "value"}, helmKubePromChartOverridings)
}

func TestKubePrometheusComponentOverridingsWithVersionAndHelmKubePromChartOverridings(t *testing.T) {
	params := stack.ComponentOverrides{
		"version":                      "v1.0.0",
		"helmKubePromChartOverridings": map[string]any{"key": "value"},
	}
	version, helmKubePromChartOverridings := setKubePrometheusComponentOverridings(params)
	assert.Equal(t, "v1.0.0", version)
	assert.Equal(t, map[string]any{"key": "value"}, helmKubePromChartOverridings)
}
