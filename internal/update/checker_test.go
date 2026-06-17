package update

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestNormalizeVersion(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"v0.5.0", "0.5.0"},
		{"0.5.0", "0.5.0"},
		{" v1.2.3 ", "1.2.3"},
		{"", ""},
	}
	for _, tc := range tests {
		if got := normalizeVersion(tc.in); got != tc.want {
			t.Errorf("normalizeVersion(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestIsNewer(t *testing.T) {
	tests := []struct {
		latest  string
		current string
		want    bool
	}{
		{"0.6.0", "0.5.0", true},
		{"0.5.0", "0.5.0", false},
		{"0.4.0", "0.5.0", false},
		{"1.0.0", "0.9.9", true},
	}
	for _, tc := range tests {
		if got := isNewer(tc.latest, tc.current); got != tc.want {
			t.Errorf("isNewer(%q, %q) = %v, want %v", tc.latest, tc.current, got, tc.want)
		}
	}
}

func TestParseGitHubRelease(t *testing.T) {
	raw := `{"tag_name":"v0.5.0","html_url":"https://github.com/PaulUno777/harmoniz/releases/tag/v0.5.0"}`
	var release githubRelease
	if err := json.Unmarshal([]byte(raw), &release); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if release.TagName != "v0.5.0" {
		t.Fatalf("tag_name = %q", release.TagName)
	}
	if normalizeVersion(release.TagName) != "0.5.0" {
		t.Fatalf("normalized = %q", normalizeVersion(release.TagName))
	}
}

func TestDownloadURLForPlatform(t *testing.T) {
	if got := downloadURLForPlatform("darwin"); got == "" || !strings.Contains(got, "harmoniz.app.zip") {
		t.Fatalf("darwin url = %q", got)
	}
	if got := downloadURLForPlatform("windows"); got == "" || !strings.Contains(got, "harmoniz-amd64-installer.exe") {
		t.Fatalf("windows url = %q", got)
	}
	if got := downloadURLForPlatform("linux"); got == "" || !strings.Contains(got, "harmoniz-linux-amd64.tar.gz") {
		t.Fatalf("linux url = %q", got)
	}
}
