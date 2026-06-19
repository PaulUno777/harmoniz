package update

import (
	"fmt"
	"net/http"
	"net/url"
	"runtime"
	"strings"
	"time"

	"golang.org/x/mod/semver"

	"harmoniz/internal/logger"
	"harmoniz/internal/version"
)

const (
	githubOwner     = "PaulUno777"
	githubRepo      = "harmoniz"
	latestRelease   = "https://github.com/PaulUno777/harmoniz/releases/latest"
	releaseTagMarker = "/releases/tag/"
)

// Result describes the outcome of a release check.
type Result struct {
	CurrentVersion  string `json:"currentVersion"`
	LatestVersion   string `json:"latestVersion"`
	UpdateAvailable bool   `json:"updateAvailable"`
	ReleasePageURL  string `json:"releasePageUrl"`
	DownloadURL     string `json:"downloadUrl"`
}

// Check resolves the latest GitHub release via the /releases/latest redirect
// (avoids REST API rate limits for unauthenticated clients).
func Check() (Result, error) {
	current := strings.TrimPrefix(strings.TrimSpace(version.Version), "v")
	if current == "" {
		err := fmt.Errorf("app version is not set")
		logger.Log.Error("Update check failed", "error", err)
		return Result{}, err
	}

	logger.Log.Info("Checking for updates", "current", current, "platform", runtime.GOOS)

	latest, releasePageURL, err := fetchLatestReleaseViaRedirect(current)
	if err != nil {
		logger.Log.Error("Update check failed", "stage", "redirect", "url", latestRelease, "error", err)
		return Result{}, err
	}

	result := Result{
		CurrentVersion:  current,
		LatestVersion:   latest,
		UpdateAvailable: isNewer(latest, current),
		ReleasePageURL:  releasePageURL,
		DownloadURL:     downloadURLForPlatform(runtime.GOOS),
	}
	logger.Log.Info(
		"Update check completed",
		"current", result.CurrentVersion,
		"latest", result.LatestVersion,
		"update_available", result.UpdateAvailable,
		"download_url", result.DownloadURL,
	)
	return result, nil
}

func fetchLatestReleaseViaRedirect(currentVersion string) (latestVersion, releasePageURL string, err error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	req, err := http.NewRequest(http.MethodHead, latestRelease, nil)
	if err != nil {
		return "", "", err
	}
	req.Header.Set("User-Agent", fmt.Sprintf("harmoniz/%s", currentVersion))

	resp, err := client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("failed to fetch latest release: %w", err)
	}
	defer resp.Body.Close()

	location := resp.Header.Get("Location")
	if location == "" {
		return "", "", fmt.Errorf("github returned %d without a release redirect", resp.StatusCode)
	}

	return parseReleaseFromURL(location)
}

func parseReleaseFromURL(rawURL string) (latestVersion, releasePageURL string, err error) {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return "", "", fmt.Errorf("invalid release redirect URL: %w", err)
	}

	idx := strings.LastIndex(parsed.Path, releaseTagMarker)
	if idx == -1 {
		return "", "", fmt.Errorf("release redirect URL has no tag: %s", rawURL)
	}

	tag := parsed.Path[idx+len(releaseTagMarker):]
	latest := normalizeVersion(tag)
	if latest == "" {
		return "", "", fmt.Errorf("release redirect URL has no valid version tag: %s", rawURL)
	}

	pageURL := parsed.Scheme + "://" + parsed.Host + parsed.Path
	return latest, pageURL, nil
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
