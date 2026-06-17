package version

// Version is the app semver without a "v" prefix.
// Keep in sync with wails.json info.productVersion.
// Release builds override via: -ldflags "-X harmoniz/internal/version.Version=x.y.z"
var Version = "0.6.0"
