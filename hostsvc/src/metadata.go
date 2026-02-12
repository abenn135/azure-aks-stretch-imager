package src

type DiskVersion struct {
	Name       string `json:"name"`
	Version    string `json:"version"`
	URL        string `json:"url"`
	ResourceID string `json:"resourceId"`
}

type Metadata struct {
	Current  DiskVersion `json:"current"`
	Next     DiskVersion `json:"next"`
	Previous DiskVersion `json:"previous"`
}
