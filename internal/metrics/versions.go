// SPDX-License-Identifier: Apache-2.0

package metrics

var (
	defaultVersions Versions
)

type Versions struct {
	Version  string
	CommitID string
}

// UpdateDefaultVersions re-assigns defaults upon init
func UpdateDefaultVersions(version, commitid string) {
	defaultVersions.Version = version
	defaultVersions.CommitID = commitid
}

// GetVersions returns the current default versions
func GetVersions() Versions {
	return defaultVersions
}
