package utility

import "testing"

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		name     string
		version1 string
		version2 string
		want     int
	}{
		{
			name:     "equal versions",
			version1: "v1.0.0",
			version2: "v1.0.0",
			want:     0,
		},
		{
			name:     "first version greater (major)",
			version1: "v2.0.0",
			version2: "v1.0.0",
			want:     1,
		},
		{
			name:     "first version lesser (major)",
			version1: "v1.0.0",
			version2: "v2.0.0",
			want:     -1,
		},
		{
			name:     "first version greater (minor)",
			version1: "v1.2.0",
			version2: "v1.1.0",
			want:     1,
		},
		{
			name:     "first version lesser (minor)",
			version1: "v1.1.0",
			version2: "v1.2.0",
			want:     -1,
		},
		{
			name:     "first version greater (patch)",
			version1: "v1.1.2",
			version2: "v1.1.1",
			want:     1,
		},
		{
			name:     "first version lesser (patch)",
			version1: "v1.1.1",
			version2: "v1.1.2",
			want:     -1,
		},
		{
			name:     "pre-release version vs release",
			version1: "v1.0.0-alpha",
			version2: "v1.0.0",
			want:     -1,
		},
		{
			name:     "with and without v prefix",
			version1: "v1.0.0",
			version2: "v1.0.0",
			want:     0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CompareVersions(tt.version1, tt.version2)
			if got != tt.want {
				t.Errorf("CompareVersions() = %v, want %v", got, tt.want)
			}
		})
	}
}
