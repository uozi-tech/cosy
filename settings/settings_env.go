package settings

import (
	"log"
	"strings"

	"github.com/caarlos0/env/v11"
)

var envPrefix = ""

func SetEnvPrefix(prefix string) {
	envPrefix = prefix
}

func parseEnv(ptr interface{}, prefix string) {
	err := env.ParseWithOptions(ptr, env.Options{
		Prefix:                envPrefix + prefix,
		UseFieldNameByDefault: true,
	})

	if err != nil {
		log.Fatalf("settings.parseEnv: %v\n", err)
	}
}

func parseAllEnv() {
	for name, ptr := range sections.AllFromFront() {
		parseEnv(ptr, strings.ToUpper(name)+"_")
	}
}
