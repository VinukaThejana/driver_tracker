// Package env is used to load enviroment variables with proper validation
package env

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/VinukaThejana/go-utils/logger"
	"github.com/spf13/viper"
)

type Env struct {
	Port int `mapstructure:"PORT" validate:"required"`
}

func (e *Env) Load(path ...string) {
	configPath := "."
	configFile := ".env"

	if len(path) > 2 {
		logger.Errorf(fmt.Errorf("invalid set of parameters are provided"))
	}

	if len(path) > 0 {
		if len(path) == 2 {
			configFile = path[1]
		}
		configPath = path[0]

		if strings.HasSuffix(path[0], "/") {
			configFile = fmt.Sprintf("%s%s", configPath, configFile)
		} else {
			configFile = fmt.Sprintf("%s/%s", configPath, configFile)
		}
	}

	_, err := os.Stat(configFile)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			logFatal(err)
		}
	} else {
		viper.AddConfigPath(configPath)
		viper.SetConfigFile(configFile)
	}

	viper.AutomaticEnv()

	logFatal(viper.ReadInConfig())
	logFatal(viper.Unmarshal(&e))

	logger.Validatef(e)
}

func logFatal(err error) {
	if err != nil {
		logger.Errorf(err)
	}
}
