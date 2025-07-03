package config

import (
	"fmt"
	"os"
	"path/filepath"
)

var ConfigPath string

func init() {
	configDir, err := os.UserConfigDir()
	if err != nil {
		fmt.Println("Could not determine config dir: ", err)
		os.Exit(1)
	}

	ConfigPath = filepath.Join(configDir, "wrong-answer-client", "config.yaml")

	os.MkdirAll(filepath.Dir(ConfigPath), 0755)

	if _, err := os.Stat(ConfigPath); os.IsNotExist(err) {
		defaultContent := []byte(`username: 
server: ws://localhost:8080/ws
auto_update: true
`)
		os.WriteFile(ConfigPath, defaultContent, 0644)
	}
}
