package update

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"strings"
	"time"

	"golang.org/x/mod/semver"

	"harmoniz/internal/version"
)

const (
	githubOwner = "PaulUno777"
	githubRepo  = "harmoniz"
	latestAPI   = "https://api.github.com/repos/PaulUno777/harmoniz/releases/latest"
)

// Result describes the outcome of a release check.
type Result struct {
	CurrentVersion  string `json:"currentVersion"`
	LatestVersion   string `json:"latestVersion"`
	UpdateAvailable bool   `json:"updateAvailable"`
	ReleasePageURL  string `json:"releasePageUrl"`
	DownloadURL     string `json:"downloadUrl"`
}

type githubRelease struct {
	TagName string `json:"tag_name"`
	HTMLURL string `json:"html_url"`
}

// Check fetches the latest GitHub release and compares it to the running app version.
func Check() (Result, error) {
	current := strings.TrimPrefix(strings.TrimSpace(version.Version), "v")
	if current == "" {
		return Result{}, fmt.Errorf("app version is not set")
	}

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest(http.MethodGet, latestAPI, nil)
	if err != nil {
		return Result{}, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", fmt.Sprintf("harmoniz/%s", current))

	resp, err := client.Do(req)
	if err != nil {
		return Result{}, fmt.Errorf("failed to fetch latest release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return Result{}, fmt.Errorf("github API returned %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return Result{}, fmt.Errorf("failed to parse release response: %w", err)
	}

	latest := normalizeVersion(release.TagName)
	if latest == "" {
		return Result{}, fmt.Errorf("release has no valid version tag")
	}

	return Result{
		CurrentVersion:  current,
		LatestVersion:   latest,
		UpdateAvailable: isNewer(latest, current),
		ReleasePageURL:  release.HTMLURL,
		DownloadURL:     downloadURLForPlatform(runtime.GOOS),
	}, nil
}

func normalizeVersion(tag string) string {
	tag = strings.TrimSpace(tag)
	return strings.TrimPrefix(tag, "v")
}

func isNewer(latest, current string) bool {
	return semver.Compare(semverTag(latest), semverTag(current)) > 0
}

func semverTag(v string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return "v0.0.0"
	}
	if !strings.HasPrefix(v, "v") {
		return "v" + v
	}
	return v
}

func downloadURLForPlatform(goos string) string {
	base := fmt.Sprintf("https://github.com/%s/%s/releases/latest/download/", githubOwner, githubRepo)
	switch goos {
	case "darwin":
		return base + "harmoniz.app.zip"
	case "windows":
		return base + "harmoniz-amd64-installer.exe"
	default:
		return base + "harmoniz-linux-amd64.tar.gz"
	}
}
