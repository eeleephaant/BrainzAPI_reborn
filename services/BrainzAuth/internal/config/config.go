package config

import (
	"github.com/ilyakaznacheev/cleanenv"
)

type (
	Config struct {
		Postgres PostgresConfig `env-prefix:"DB_"`
		Auth     AuthConfig
		Redis    RedisConfig `env-prefix:"REDIS_"`
		App      AppConfig   `env-prefix:"APP_"`
	}

	AppConfig struct {
		Address     string `env:"ADDRESS, required"`
		Port        int    `env:"PORT, required"`
		Environment string `env:"ENV"`
	}

	AuthConfig struct {
		Pepper        string `env:"PEPPER, required"`
		Argon2Memory  int    `env:"ARGON2_MEMORY, required"`
		Argon2Time    int    `env:"ARGON2_TIME, required"`
		Argon2Threads int    `env:"ARGON2_THREADS, required"`
	}

	PostgresConfig struct {
		Host     string `env:"HOST, required"`
		Port     string `env:"PORT, required"`
		User     string `env:"USER, required"`
		Name     string `env:"NAME, required"`
		Password string `env:"PASSWORD, required"`
		PoolMax  int32  `env:"POOL_MAX,required"`
	}

	RedisConfig struct {
		Host     string `env:"HOST,required"`
		Password string `env:"PASSWORD,required"`
	}
)

func MustLoad() *Config {
	var cfg Config
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		panic("cannot read env: " + err.Error())
	}
	return &cfg
}
