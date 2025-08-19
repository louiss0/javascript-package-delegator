// Package env provides utilities for managing Go environment mode detection
// and executing mode-specific functionality.
package env

import (
	"github.com/louiss0/javascript-package-delegator/build_info"
)

type GoEnv struct {
	goEnv string
}

func NewGoEnv() GoEnv {
	return GoEnv{build_info.GO_MODE.String()}
}

// Mode returns the current Go environment mode string (e.g., "production", "development").
// Per naming rules, avoid a Get prefix for simple getters.
func (env GoEnv) Mode() string {
	return env.goEnv
}

func (env GoEnv) IsDebugMode() bool {
	return env.goEnv == "debug"
}

func (env GoEnv) IsDevelopmentMode() bool {
	return env.goEnv == "development" || env.goEnv == ""
}

func (env GoEnv) IsProductionMode() bool {
	return env.goEnv == "production"
}

func (env GoEnv) ExecuteIfModeIsProduction(cb func()) {
	if env.IsProductionMode() {
		cb()
	}
}
