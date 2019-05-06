package version

// Default build-time variable.
// These values can (should) be overridden via ldflags when built with
// `make`
var (
	Version   = "unknown-version"
	GoVersion = "unknown-goversion"
	GitCommit = "unknown-commit"
	BuildTime = "unknown-buildtime"
	Os        = "unknown-os"
	Arch      = "unknown-arch"
)
