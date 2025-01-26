package apps

import (
	"testing"

	"github.com/ksctl/ksctl/v2/pkg/utilities"
)

func TestGetVersionIfItsNotNilAndLatest(t *testing.T) {
	tests := []struct {
		name        string
		version     *string
		defaultVer  string
		expectedVer string
	}{
		{
			name:        "version is nil",
			version:     nil,
			defaultVer:  "v0.0.1",
			expectedVer: "v0.0.1",
		},
		{
			name:        "version is latest",
			version:     utilities.Ptr("latest"),
			defaultVer:  "v0.0.1",
			expectedVer: "v0.0.1",
		},
		{
			name:        "version is not latest",
			version:     utilities.Ptr("v1.0.0"),
			defaultVer:  "v0.0.1",
			expectedVer: "v1.0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualVer := GetVersionIfItsNotNilAndLatest(tt.version, tt.defaultVer)
			if actualVer != tt.expectedVer {
				t.Errorf("expected version %s, got %s", tt.expectedVer, actualVer)
			}
		})
	}
}
