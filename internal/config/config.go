package config

import (
	"encoding/json"
	"log"
	"os"
)

const configFileName = ".gatorconfig.json"

type Config struct {
	DBURL    string `json:"db_url"`
	Username string `json:"current_user_name"`
}

func Read() Config {
	var cfg Config
	configFilePath, err := getConfigFilePath()
	if err != nil {
		log.Printf("Error retrieving config file path: %s\n", err)
		return cfg
	}
	f, err := os.Open(configFilePath)
	if err != nil {
		log.Printf("Error opening config file for reading: %s\n", err)
		return cfg
	}
	defer f.Close()
	dec := json.NewDecoder(f)
	if err := dec.Decode(&cfg); err != nil {
		log.Println(err)
	}
	return cfg
}

func (cfg *Config) SetUser(user string) {
	cfg.Username = user
	err := write(*cfg)
	if err != nil {
		log.Printf("Error setting username: %s\n", err)
	}
}

func getConfigFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return homeDir + "/" + configFileName, nil
}

func write(cfg Config) error {
	configFilePath, err := getConfigFilePath()
	if err != nil {
		log.Printf("Error retriving config file path: %s\n", err)
		return err
	}
	f, err := os.Create(configFilePath)
	if err != nil {
		log.Printf("Error opening config file for writing: %s\n", err)
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	if err := enc.Encode(cfg); err != nil {
		log.Println(err)
		return err
	}
	return nil
}
