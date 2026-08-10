package utility

import (
	"strings"

	"golang.org/x/mod/semver"
)

func CompareVersions(version1, version2 string) int {
	// Ensure versions start with "v" prefix for semver.Compare
	v1 := normalizeVersion(version1)
	v2 := normalizeVersion(version2)

	return semver.Compare(v1, v2)
}

func normalizeVersion(version string) string {
	if version == "" {
		return ""
	}

	// Add "v" prefix if it doesn't exist
	if !strings.HasPrefix(version, "v") {
		return "v" + version
	}

	return version
}
