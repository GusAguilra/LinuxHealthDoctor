package version

var (
	Version   = "0.1.0-dev"
	Commit    = "none"
	Date      = "unknown"
	BuiltBy   = "unknown"
)

func Info() string {
	return Version
}

func FullInfo() string {
	return "lhd v" + Version + " (commit: " + Commit + ", built: " + Date + ")"
}
