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
	goEnv       string
	goEnvExists bool
}

func NewGoEnv() (GoEnv, error) {
	goEnv, ok := getGoEnv()

	if goEnv != "" && !lo.Contains(allowedModes, goEnv) {

		return GoEnv{}, fmt.Errorf("Wrong go mode the only allowed modes are %v", allowedModes)
	}

	return GoEnv{goEnv, ok}, nil
}

func (e GoEnv) GetgoEnv() string {
	return e.goEnv
}

func (e GoEnv) IsDebugMode() bool {

	return e.goEnv == "debug"
}

func (env GoEnv) IsDevelopmentMode() bool {
	return env.goEnv == "development"
}

func (env GoEnv) IsProductionMode() bool {
	return env.goEnv == "production" || env.goEnvExists == false
}

func (env GoEnv) ExecuteIfModeIsProduction(cb func()) {

	if env.IsProductionMode() {
		cb()

	}
}
