package version

import (
	"fmt"
)

var (
	// Major is the major version number.
	Major = 0

	// Minor is the minor version number.
	Minor = 2

	// Patch is the patch version number.
	Patch = 2

	// Commit is the commit the application was built on.
	Commit = ""
)

// Version returns the application version as a properly formed string per the
// semantic versioning 2.0.0 spec (http://semver.org/) and the commit it was
// built on.
func Version() string {
	semver := fmt.Sprintf("%d.%d.%d", Major, Minor, Patch)
	if Commit == "" {
		return semver
	}

	return Commit
}
