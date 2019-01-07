package config

import (
	"os"
	"tp-project-db/consts"
)

func Load() {
	setEnvVar("SERVER_HOST", "0.0.0.0")
	setEnvVar("SERVER_PORT", "5000")

	/*setEnvVar("PGHOST", "127.0.0.1")
	setEnvVar("PGPORT", "5432")
	setEnvVar("PGDATABASE", "forum")
	setEnvVar("PGUSER", "postgres")
	setEnvVar("PGPASSWORD", consts.EmptyString)*/

	setEnvVar("PGHOST", "127.0.0.1")
	setEnvVar("PGPORT", "5432")
	setEnvVar("PGDATABASE", "forum")
	setEnvVar("PGUSER", "user")
	setEnvVar("PGPASSWORD", "password")
}

func getEnvVar(name, defaultValue string) string {
	if v := os.Getenv(name); v != consts.EmptyString {
		return v
	}
	return defaultValue
}

func setEnvVar(name, defaultValue string) {
	if os.Getenv(name) != consts.EmptyString {
		return
	}
	_ = os.Setenv(name, defaultValue)
}
