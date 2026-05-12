package config

import (
	"os"
	"sync/atomic"
)

type EnvVars struct {
	IsDev        bool
	ServerSecret string
	DBUrl        string
	PolkaKey     string
}

type Config struct {
	Env            EnvVars
	FileserverHits atomic.Int32
}

func Load() *Config {
	env := EnvVars{
		IsDev:        os.Getenv("PLATFORM") == "dev",
		ServerSecret: os.Getenv("SERVER_SECRET"),
		DBUrl:        os.Getenv("DB_URL"),
		PolkaKey:     os.Getenv("POLKA_KEY"),
	}

	cfg := &Config{
		Env: env,
	}

	return cfg
}
