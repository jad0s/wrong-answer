package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

var ConfigPath string
var Config map[string]string

func init() {
	configDir, err := os.UserConfigDir()
	if err != nil {
		fmt.Println("Could not determine config dir: ", err)
		os.Exit(1)
	}

	ConfigPath = filepath.Join(configDir, "wrong-answer-server", "config.yaml")

	os.MkdirAll(filepath.Dir(ConfigPath), 0755)

	if _, err := os.Stat(ConfigPath); os.IsNotExist(err) {
		defaultContent := []byte(`port: 8080
answer_timer: 20
vote_timer: 180
auto_update: false
`)
		os.WriteFile(ConfigPath, defaultContent, 0644)
	}
}

func LoadConfig(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("could not read config file: %w", err)
	}

	err = yaml.Unmarshal(data, &Config)
	if err != nil {
		return fmt.Errorf("could not parse yaml: %w", err)
	}

	return nil

}
