package analyzer

import (
	"encoding/json"
	"log"
	"os"
)

type BackendConfig struct {
	Datastore   string
	Storage     string
	Environment string
}

func LoadConfig(configFile string) BackendConfig {
	fh, err := os.Open(configFile)
	if err != nil {
		log.Printf("Unable to open configuration: %s", configFile)
		panic(err)
	}
	defer fh.Close()

	config := BackendConfig{}
	dec := json.NewDecoder(fh)
	if err := dec.Decode(&config); err != nil {
		log.Printf("Unable to read configuration: %s", configFile)
		panic(err)
	}

	return config
}
