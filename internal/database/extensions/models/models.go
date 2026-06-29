package models

type ExtensionInfo struct {
	Name             string `json:"name"`
	Version          string `json:"version"`
	Description      string `json:"description"`
	Installed        bool   `json:"installed"`
	InstalledVersion string `json:"installed_version"`
}
