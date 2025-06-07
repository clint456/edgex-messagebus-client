package messagebus

import (
	"runtime"
	"strings"
	"testing"
)

func TestGetVersion(t *testing.T) {
	version := GetVersion()

	if version.Version == "" {
		t.Error("Version should not be empty")
	}

	if version.GoVersion == "" {
		t.Error("GoVersion should not be empty")
	}

	if version.Platform == "" {
		t.Error("Platform should not be empty")
	}

	// Check that GoVersion matches runtime
	if version.GoVersion != runtime.Version() {
		t.Errorf("Expected GoVersion %s, got %s", runtime.Version(), version.GoVersion)
	}

	// Check platform format
	expectedPlatform := runtime.GOOS + "/" + runtime.GOARCH
	if version.Platform != expectedPlatform {
		t.Errorf("Expected Platform %s, got %s", expectedPlatform, version.Platform)
	}
}

func TestGetVersionString(t *testing.T) {
	versionString := GetVersionString()

	if !strings.HasPrefix(versionString, "v") {
		t.Error("Version string should start with 'v'")
	}

	if !strings.Contains(versionString, Version) {
		t.Errorf("Version string should contain version %s", Version)
	}
}

func TestVersionInfoString(t *testing.T) {
	version := GetVersion()
	versionString := version.String()

	if versionString == "" {
		t.Error("Version string should not be empty")
	}

	// Check that the string contains expected components
	expectedComponents := []string{
		"EdgeX MessageBus Client",
		version.Version,
		version.GoVersion,
		version.Platform,
	}

	for _, component := range expectedComponents {
		if !strings.Contains(versionString, component) {
			t.Errorf("Version string should contain '%s', got: %s", component, versionString)
		}
	}
}

func TestVersionConstants(t *testing.T) {
	if Version == "" {
		t.Error("Version constant should not be empty")
	}

	// Version should follow semantic versioning pattern (basic check)
	parts := strings.Split(Version, ".")
	if len(parts) < 2 {
		t.Errorf("Version should have at least major.minor format, got: %s", Version)
	}
}
