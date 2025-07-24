package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	Port            string   `json:"port"`
	AllowedExt      []string `json:"allowed_extensions"`
	MaxTasks        int      `json:"max_tasks"`
	MaxFilesPerTask int      `json:"max_files_per_task"`
}

func Load() (*Config, error) {
	// Значения по умолчанию
	cfg := &Config{
		Port:            "8080",
		AllowedExt:      []string{".pdf", ".jpeg"},
		MaxTasks:        3,
		MaxFilesPerTask: 3,
	}

	// Чтение config.json
	data, err := os.ReadFile("config.json")
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, err
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
