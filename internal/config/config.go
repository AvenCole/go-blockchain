package config

// Config defines the initialization-stage runtime settings shared by the CLI
// and later blockchain modules.
type Config struct {
	ProjectName string
	DataDir     string
	DefaultPort int
	LogLevel    string
	NetworkMode string
}

// Default returns the initialization-safe defaults agreed in Plan 1.
func Default() Config {
	return Config{
		ProjectName: "go-blockchain",
		DataDir:     "./data",
		DefaultPort: 3000,
		LogLevel:    "info",
		NetworkMode: "local",
	}
}
