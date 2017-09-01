package changelog

// ChangeLog is a type for general configuration for producing a changelog
type ChangeLog struct {
	Repo         string `toml:"repo"`
	Version      string `toml:"version"`
	PatchVersion bool   `toml:"patch_ver"`
	Subtitle     string `toml:"subtitle"`
}
