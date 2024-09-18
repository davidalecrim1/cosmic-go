package env

import "os"

func GetEnvOrSetDefault(key string, defaultValue string) string {
	env := os.Getenv(key)

	if env == "" {
		env = defaultValue
	}

	return env
}
