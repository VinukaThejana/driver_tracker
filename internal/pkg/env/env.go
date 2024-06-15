// Package env is used to load enviroment variables with proper validation
package env

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/VinukaThejana/go-utils/logger"
	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/lib"
	"github.com/spf13/viper"
)

// Env contains the env schema
type Env struct {
	KafkaUsername       string        `mapstructure:"KAFKA_USERNAME" validate:"required"`
	KafkaPassword       string        `mapstructure:"KAFKA_PASSWORD" validate:"required"`
	KafkaBroker         string        `mapstructure:"KAFKA_BROKER" validate:"required"`
	KafkaTopic          string        `mapstructure:"KAFKA_TOPIC" validate:"required"`
	KafkaRestURL        string        `mapstructure:"KAFKA_REST_URL" validate:"required,url"`
	RedisDBURL          string        `mapstructure:"REDIS_DB_URL" validate:"required"`
	Domain              string        `mapstructure:"DOMAIN" validate:"required,url"`
	Host                string        `mapstructure:"HOST" validate:"required"`
	APIDoc              string        `mapstructure:"API_DOC" validate:"required,url"`
	BookingTokenSecret  string        `mapstructure:"BOOKING_TOKEN_SECRET" validate:"required"`
	PartitionManagerKey string        `mapstructure:"PARTITION_MANAGER_KEY" validate:"required"`
	Topic               string        `mapstructure:"TOPIC" validate:"required"`
	BookingTokenExpires time.Duration `mapstructure:"BOOKING_TOKEN_EXPIRES_IN" validate:"required"`
	Port                int           `mapstructure:"PORT" validate:"required"`
}

// Load is a function that is used to Load environment variables
func (e *Env) Load(path ...string) {
	configPath := "."
	configFile := ".env"

	v := viper.New()

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
			lib.LogFatal(err)
		}

		lib.LogFatal(parseEnvVars(environ(), e))
	} else {
		v.AddConfigPath(configPath)
		v.SetConfigFile(configFile)

		lib.LogFatal(v.ReadInConfig())
		lib.LogFatal(v.Unmarshal(&e))
	}

	logger.Validatef(e)
}

func environ() map[string]string {
	m := make(map[string]string)
	for _, s := range os.Environ() {
		a := strings.Split(s, "=")
		m[a[0]] = a[1]
	}

	return m
}

func parseEnvVars(envMap map[string]string, e interface{}) error {
	objValue := reflect.ValueOf(e).Elem()
	objType := objValue.Type()

	for i := 0; i < objType.NumField(); i++ {
		field := objType.Field(i)
		envKey := field.Tag.Get("mapstructure")
		envValue, ok := envMap[envKey]
		if !ok {
			continue
		}

		fieldValue := objValue.Field(i)
		if !fieldValue.CanSet() {
			return fmt.Errorf("field %s is not settable", field.Name)
		}

		switch fieldValue.Kind() {
		case reflect.String:
			fieldValue.SetString(envValue)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			val, err := strconv.ParseInt(envValue, 10, 64)
			if err != nil {
				return fmt.Errorf("failed to parse %s as int: %v", envKey, err)
			}

			fieldValue.SetInt(val)
		default:
			return fmt.Errorf("unsupported type for field %s", field.Name)
		}
	}

	return nil
}
