package main

import (
	"github.com/seguidor777/portfel/internal/models"
	"os"

	"gopkg.in/yaml.v3"
)

func readConfig(path *string) (config *models.Config, err error) {
	data, err := os.ReadFile(*path)
	if err != nil {
		return nil, err
	}

	config = &models.Config{}
	err = yaml.Unmarshal(data, config)

	return
}
