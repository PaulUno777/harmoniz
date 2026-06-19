package update

import (
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

func TestParseReleaseFromURL(t *testing.T) {
	tests := []struct {
		rawURL      string
		wantVersion string
		wantPage    string
		wantErr     bool
	}{
		{
			rawURL:      "https://github.com/PaulUno777/harmoniz/releases/tag/v0.7.0",
			wantVersion: "0.7.0",
			wantPage:    "https://github.com/PaulUno777/harmoniz/releases/tag/v0.7.0",
		},
		{
			rawURL:  "https://github.com/PaulUno777/harmoniz/releases/latest",
			wantErr: true,
		},
	}
	for _, tc := range tests {
		gotVersion, gotPage, err := parseReleaseFromURL(tc.rawURL)
		if tc.wantErr {
			if err == nil {
				t.Fatalf("parseReleaseFromURL(%q) expected error", tc.rawURL)
			}
			continue
		}
		if err != nil {
			t.Fatalf("parseReleaseFromURL(%q): %v", tc.rawURL, err)
		}
		if gotVersion != tc.wantVersion {
			t.Fatalf("version = %q, want %q", gotVersion, tc.wantVersion)
		}
		if gotPage != tc.wantPage {
			t.Fatalf("page = %q, want %q", gotPage, tc.wantPage)
		}
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
