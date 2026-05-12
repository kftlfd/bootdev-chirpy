package config

import (
	"os"
)

type EnvVars struct {
	Platform     string
	ServerSecret string
	DBUrl        string
	PolkaKey     string
}

func getEnv() EnvVars {
	return EnvVars{
		Platform:     os.Getenv("PLATFORM"),
		ServerSecret: os.Getenv("SERVER_SECRET"),
		DBUrl:        os.Getenv("DB_URL"),
		PolkaKey:     os.Getenv("POLKA_KEY"),
	}
}

type Config struct {
	Env   EnvVars
	IsDev bool
}

func Load() *Config {
	env := getEnv()

	return &Config{
		Env:   env,
		IsDev: env.Platform == "dev",
	}
}
