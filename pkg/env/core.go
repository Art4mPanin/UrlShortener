package env

import (
	"fmt"
	"github.com/joho/godotenv"
	"os"
)

var initialized = false

type Env struct {
}

type KeyEnvError struct {
	key string
}

type TypeEnvError struct {
	value        interface{}
	expectedType string
}

func (e KeyEnvError) Error() string {
	return fmt.Sprintf("No such key in environment: %s, default value was not provided", e.key)
}

func (e TypeEnvError) Error() string {
	return fmt.Sprintf("Failed to convert type to expected type %s, got value: %s", e.expectedType, e.value)
}

func NewEnv(filepath *string) *Env {
	newEnv := &Env{}
	if !initialized {
		err := newEnv.LoadEnv(filepath)
		if err != nil {
			panic(fmt.Sprintf("error loading env file: %v", err))
		}
	}
	return newEnv
}

func (e *Env) LoadEnv(filepath *string) error {
	if initialized {
		return nil
	}

	if filepath == nil {
		return fmt.Errorf("filepath is nil")
	}
	_, err := os.Stat(*filepath)

	if os.IsNotExist(err) {
		return err
	}

	if err := godotenv.Load(*filepath); err != nil {
		return err
	}

	initialized = true

	return nil
}
