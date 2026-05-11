package config

import (
	"os"
	"sync/atomic"
)

type EnvVars struct {
	IsDev bool
	DBUrl string
}

type Config struct {
	Env            EnvVars
	FileserverHits atomic.Int32
}

func Load() *Config {
	env := EnvVars{
		IsDev: os.Getenv("PLATFORM") == "dev",
		DBUrl: os.Getenv("DB_URL"),
	}

	cfg := &Config{
		Env: env,
	}

	return cfg
}
