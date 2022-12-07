// Package version provides the version of packages.
package version

// swagger:model VersionResponse
type Version struct {
	BuildVersion string `json:"BuildVersion"`
	BuildType    string `json:"BuildType"`
	BuildDate    string `json:"BuildDate"`
}

func DefaultVersion() *Version {
	return &Version{
		BuildType: "dev",
	}
}
