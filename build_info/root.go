// This package defines all the build info that this repo needs
// All variables must be capitalised and must use the `BuildInfo` type
// All validation of individual types must be handled by the init function
package build_info

import (
	"fmt"
	"regexp"
	"time"

	"github.com/samber/lo"
)

type BuildInfo string

func (value BuildInfo) String() string {
	return string(value)
}

var CLI_VERSION BuildInfo
var GO_MODE BuildInfo

var currentTime = time.Now()

// It's best never to change this variable at all! It can
var BUILD_DATE = BuildInfo(currentTime.Format(time.DateOnly))

func init() {

	var initalizedBuildDate = currentTime.Format(time.DateOnly)

	var allowedModes = []string{"development", "production", "debug"}

	var semverRegex = `^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-((?:0|[1-9]\d*|[0-9]*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|[0-9]*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`

	if !lo.Contains(allowedModes, CLI_VERSION.String()) {

		panic(fmt.Sprintf("The go mode must be either %v", allowedModes))

	}

	if match := regexp.MustCompile(semverRegex).MatchString(string(GO_MODE)); !match {

		panic(fmt.Sprintf("The go mode must be a valid semver string"))
	}

	if BUILD_DATE.String() != initalizedBuildDate {
		panic(fmt.Sprintf("The build date must be the same as the current date.\nDon't change the build date"))
	}

}
