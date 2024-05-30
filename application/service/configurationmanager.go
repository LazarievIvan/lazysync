package service

import (
	"gopkg.in/yaml.v3"
	"log"
	"os"
)

const ConfigFile = "config.yaml"

type AppConfiguration struct {
	Mode                 string `yaml:"mode"`
	Username             string `yaml:"username"`
	Module               string `yaml:"module"`
	ModuleSpecificConfig any    `yaml:"config"`
}

func SaveConfiguration(configuration *AppConfiguration) {
	yamlContents, err := yaml.Marshal(configuration)
	if err != nil {
		panic(err)
	}
	err = os.WriteFile(ConfigFile, yamlContents, 0644)
	if err != nil {
		panic(err)
	}
}

func LoadConfiguration() *AppConfiguration {
	var config AppConfiguration
	yamlFile, err := os.ReadFile(ConfigFile)
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
	}
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}
	return &config
}
