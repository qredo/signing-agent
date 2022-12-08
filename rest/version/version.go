// Package version provides the version of packages.
package version

// swagger:model VersionResponse
type Version struct {
	BuildVersion string `json:"buildVersion"`
	BuildType    string `json:"buildType"`
	BuildDate    string `json:"buildDate"`
}

func DefaultVersion() *Version {
	return &Version{
		BuildType: "dev",
	}
}
