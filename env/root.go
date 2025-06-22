 package env

import (
	"fmt"
	"os"

	"github.com/samber/lo"
)

const _GO_MODE_ENV_KEY = "GO_MODE"

func getGoEnv() (string, bool) {
	return os.LookupEnv(_GO_MODE_ENV_KEY)
}

var allowedModes = []string{"development", "production", "debug"}

type GoEnv struct {
	goMode       string
	goModeExists bool
}

func NewGoEnv() (GoEnv, error) {
	goMode, ok := getGoEnv()

	if goMode != "" && !lo.Contains(allowedModes, goMode) {

		return GoEnv{}, fmt.Errorf("Wrong go mode the only allowed modes are %v", allowedModes)
	}

	return GoEnv{goMode, ok}, nil
}

func (e GoEnv) GetGoMode() string {
	return e.goMode
}

func (e GoEnv) IsDebugMode() bool {

	return e.goMode == "debug"
}

func (env GoEnv) IsDevelopmentMode() bool {
	return env.goMode == "development"
}

func (env GoEnv) IsProductionMode() bool {
	return env.goMode == "production" || env.goModeExists == false
}

func (env GoEnv) ExecuteIfModeIsProduction(cb func()) {

	if env.IsProductionMode() {
		cb()

	}
}
