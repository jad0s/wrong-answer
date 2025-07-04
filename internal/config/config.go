package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

var ConfigPath string
var QuestionsPath string

var Config map[string]string

type QuestionPair struct {
	Normal   string `json:"normal"`
	Impostor string `json:"impostor"`
}

var Questions []QuestionPair

func init() {
	configDir, err := os.UserConfigDir()
	if err != nil {
		fmt.Println("Could not determine config dir:", err)
		os.Exit(1)
	}

	ConfigPath = filepath.Join(configDir, "wrong-answer-server", "config.yaml")
	QuestionsPath = filepath.Join(configDir, "wrong-answer-server", "questions.json")

	os.MkdirAll(filepath.Dir(ConfigPath), 0755)

	// Write default config.yaml if not exists
	if _, err := os.Stat(ConfigPath); os.IsNotExist(err) {
		defaultContent := []byte(`port: "8080"
answer_timer: "20"
vote_timer: "180"
auto_update: "false"
`)
		os.WriteFile(ConfigPath, defaultContent, 0644)
	}

	// Write default questions.json if not exists
	if _, err := os.Stat(QuestionsPath); os.IsNotExist(err) {
		defaultQuestions := []QuestionPair{
			{Normal: "What's your favorite fruit?", Impostor: "What's your favorite cheese?"},
			{Normal: "What do you eat for breakfast?", Impostor: "What do you eat for dinner?"},
		}
		data, _ := json.MarshalIndent(defaultQuestions, "", "  ")
		os.WriteFile(QuestionsPath, data, 0644)
	}
}

func LoadConfig(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("could not read config file: %w", err)
	}

	return yaml.Unmarshal(data, &Config)
}

func LoadQuestions(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("could not read questions file: %w", err)
	}

	return json.Unmarshal(data, &Questions)
}
