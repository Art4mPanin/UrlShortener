package env

import (
	"fmt"
	"log"
	"os"
	"strconv"
)

type EnvGetter interface {
	GetEnv(key string, defaultValue ...interface{}) (interface{}, error)
	GetEnvAsString(key string, defaultValue ...string) (string, error)
	GetEnvAsInt(key string, defaultValue ...int) (int, error)
	GetEnvAsBool(key string, defaultValue ...bool) (bool, error)
}

type EnvGetterPanic interface {
	EnvGetter
	GetEnvPanic(key string, defaultValue ...interface{}) interface{}
	GetEnvAsStringPanic(key string, defaultValue ...string) string
	GetEnvAsIntPanic(key string, defaultValue ...int) int
	GetEnvAsBoolPanic(key string, defaultValue ...bool) bool
}

func (e *Env) GetEnv(key string, defaultValue ...interface{}) (interface{}, error) {
	val, exists := os.LookupEnv(key)

	if !exists {
		if defaultValue != nil {
			log.Printf("WARN: GetEnv failed, returning default value: %v", defaultValue[0])
			return defaultValue[0], nil
		}
		return "", &KeyEnvError{key: key}
	}
	return val, nil
}

func (e *Env) GetEnvAsString(key string, defaultValue ...string) (string, error) {
	val, err := e.GetEnv(key)

	if err != nil {
		if defaultValue != nil {
			log.Printf("WARN: GetEnvAsString failed, returning default value: %v", defaultValue[0])
			return defaultValue[0], nil
		}
		return "", &KeyEnvError{key: key}
	}
	return val.(string), nil
}

func (e *Env) GetEnvAsInt(key string, defaultValue ...int) (int, error) {
	val, err := e.GetEnv(key)
	if err != nil {
		if defaultValue != nil {
			return defaultValue[0], nil
		}
		return 0, err
	}
	intVal, err := strconv.Atoi(val.(string))
	if err != nil {
		if defaultValue != nil {
			log.Printf("WARN: GetEnvAsInt failed, returning default value: %v", defaultValue[0])
			return defaultValue[0], nil
		}
		return 0, &TypeEnvError{value: val, expectedType: "int"}
	}
	return intVal, nil
}

func (e *Env) GetEnvAsBool(key string, defaultValue ...bool) (bool, error) {
	val, err := e.GetEnv(key)
	if err != nil {
		if defaultValue != nil {
			return defaultValue[0], nil
		}
		return false, err
	}
	boolVal, err := strconv.ParseBool(val.(string))
	if err != nil {
		if defaultValue != nil {
			log.Printf("WARN: GetEnvAsBool failed, returning default value: %v", defaultValue[0])
			return defaultValue[0], nil
		}
		return false, &TypeEnvError{value: val, expectedType: "bool"}
	}
	return boolVal, nil
}

func panicFormat(funcName string, key string, err error) {
	panic(fmt.Sprintf("Error getting key `%s` from env using func `%s`: %v", key, funcName, err))
}

func (e *Env) GetEnvPanic(key string, defaultValue ...interface{}) interface{} {
	val, err := e.GetEnv(key, defaultValue...)
	if err != nil {
		panicFormat("GetEnv", key, err)
	}
	return val
}

func (e *Env) GetEnvAsStringPanic(key string, defaultValue ...string) string {
	val, err := e.GetEnvAsString(key, defaultValue...)
	if err != nil {
		panicFormat("GetEnvAsString", key, err)
	}
	return val
}

func (e *Env) GetEnvAsIntPanic(key string, defaultValue ...int) int {
	val, err := e.GetEnvAsInt(key, defaultValue...)
	if err != nil {
		panicFormat("GetEnvAsInt", key, err)
	}
	return val
}

func (e *Env) GetEnvAsBoolPanic(key string, defaultValue ...bool) bool {
	val, err := e.GetEnvAsBool(key, defaultValue...)
	if err != nil {
		panicFormat("GetEnvAsBool", key, err)
	}
	return val
}
