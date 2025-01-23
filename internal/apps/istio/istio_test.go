package istio

import (
	"github.com/ksctl/ksctl/pkg/apps/stack"
	"github.com/ksctl/ksctl/pkg/poller"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	poller.InitSharedGithubReleaseFakePoller(func(org, repo string) ([]string, error) {
		vers := []string{"v0.0.1"}

		switch org + " " + repo {
		case "istio istio":
			vers = append(vers, "v1.22.4")
		}

		sort.Slice(vers, func(i, j int) bool {
			return vers[i] > vers[j]
		})

		for i := range vers {
			if strings.HasPrefix(vers[i], "v") {
				vers[i] = strings.TrimPrefix(vers[i], "v")
			}
		}

		return vers, nil
	})
	m.Run()
}

func TestIsitoComponentOverridingsWithNilParams(t *testing.T) {
	version, helmBaseChartOverridings, helmIstiodChartOverridings, err := setIsitoComponentOverridings(nil)
	assert.Nil(t, err)
	assert.Equal(t, "1.22.4", version)
	assert.Equal(t, map[string]any{"defaultRevision": "default"}, helmBaseChartOverridings)
	assert.Nil(t, helmIstiodChartOverridings)
}

func TestIsitoComponentOverridingsWithEmptyParams(t *testing.T) {
	params := stack.ComponentOverrides{}
	version, helmBaseChartOverridings, helmIstiodChartOverridings, err := setIsitoComponentOverridings(params)
	assert.Nil(t, err)
	assert.Equal(t, "1.22.4", version)
	assert.Equal(t, map[string]any{"defaultRevision": "default"}, helmBaseChartOverridings)
	assert.Nil(t, helmIstiodChartOverridings)
}

func TestIsitoComponentOverridingsWithVersionOnly(t *testing.T) {
	params := stack.ComponentOverrides{
		"version": "v1.0.0",
	}
	version, helmBaseChartOverridings, helmIstiodChartOverridings, err := setIsitoComponentOverridings(params)
	assert.Nil(t, err)
	assert.Equal(t, "v1.0.0", version)
	assert.Equal(t, map[string]any{"defaultRevision": "default"}, helmBaseChartOverridings)
	assert.Nil(t, helmIstiodChartOverridings)
}

func TestIsitoComponentOverridingsWithHelmBaseChartOverridingsOnly(t *testing.T) {
	params := stack.ComponentOverrides{
		"helmBaseChartOverridings": map[string]any{"key": "value"},
	}
	version, helmBaseChartOverridings, helmIstiodChartOverridings, err := setIsitoComponentOverridings(params)
	assert.Nil(t, err)
	assert.Equal(t, "1.22.4", version)
	assert.Equal(t, map[string]any{"key": "value"}, helmBaseChartOverridings)
	assert.Nil(t, helmIstiodChartOverridings)
}

func TestIsitoComponentOverridingsWithHelmIstiodChartOverridingsOnly(t *testing.T) {
	params := stack.ComponentOverrides{
		"helmIstiodChartOverridings": map[string]any{"key": "value"},
	}
	version, helmBaseChartOverridings, helmIstiodChartOverridings, err := setIsitoComponentOverridings(params)
	assert.Nil(t, err)
	assert.Equal(t, "1.22.4", version)
	assert.Equal(t, map[string]any{"defaultRevision": "default"}, helmBaseChartOverridings)
	assert.Equal(t, map[string]any{"key": "value"}, helmIstiodChartOverridings)
}

func TestIsitoComponentOverridingsWithVersionAndHelmBaseChartOverridings(t *testing.T) {
	params := stack.ComponentOverrides{
		"version":                  "v1.0.0",
		"helmBaseChartOverridings": map[string]any{"key": "value"},
	}
	version, helmBaseChartOverridings, helmIstiodChartOverridings, err := setIsitoComponentOverridings(params)
	assert.Nil(t, err)
	assert.Equal(t, "v1.0.0", version)
	assert.Equal(t, map[string]any{"key": "value"}, helmBaseChartOverridings)
	assert.Nil(t, helmIstiodChartOverridings)
}

func TestIsitoComponentOverridingsWithVersionAndHelmIstiodChartOverridings(t *testing.T) {
	params := stack.ComponentOverrides{
		"version":                    "v1.0.0",
		"helmIstiodChartOverridings": map[string]any{"key": "value"},
	}
	version, helmBaseChartOverridings, helmIstiodChartOverridings, err := setIsitoComponentOverridings(params)
	assert.Nil(t, err)
	assert.Equal(t, "v1.0.0", version)
	assert.Equal(t, map[string]any{"defaultRevision": "default"}, helmBaseChartOverridings)
	assert.Equal(t, map[string]any{"key": "value"}, helmIstiodChartOverridings)
}

func TestIsitoComponentOverridingsWithAllParams(t *testing.T) {
	params := stack.ComponentOverrides{
		"version":                    "1.0.0",
		"helmBaseChartOverridings":   map[string]any{"baseKey": "baseValue"},
		"helmIstiodChartOverridings": map[string]any{"istiodKey": "istiodValue"},
	}
	version, helmBaseChartOverridings, helmIstiodChartOverridings, err := setIsitoComponentOverridings(params)
	assert.Nil(t, err)
	assert.Equal(t, "1.0.0", version)
	assert.Equal(t, map[string]any{"baseKey": "baseValue"}, helmBaseChartOverridings)
	assert.Equal(t, map[string]any{"istiodKey": "istiodValue"}, helmIstiodChartOverridings)
}
