package config

import (
	"log"

	"github.com/tkanos/gonfig"
)

type configuration struct {
	ConnectionString   string
	Port               int
	Url                string
	CORSAllowedOrigins []string
}

func Config() configuration {
	config := configuration{}
	err := gonfig.GetConf("config/config.json", &config)

	if err != nil {
		log.Fatalf("cannot read the config file. GetConf returned: %s", err.Error())
	}
	return config
}
