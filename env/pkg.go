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

func (e GoEnv) GetgoEnv() string {
	return e.goEnv
}

func (e GoEnv) IsDebugMode() bool {

	return e.goEnv == "debug"
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
