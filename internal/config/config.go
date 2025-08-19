package config

import (
	"encoding/json"
	"fmt"
	"log"
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
	//gets configuration directory from OS. ~/.config on linux.
	configDir, err := os.UserConfigDir()
	if err != nil {
		fmt.Println("Could not determine config dir:", err)
		os.Exit(1)
	}
	//Config files stored in ~/.config/wrong-answer-server
	ConfigPath = filepath.Join(configDir, "wrong-answer-server", "config.yaml")
	QuestionsPath = filepath.Join(configDir, "wrong-answer-server", "questions.json")

	//if ~/.config/wrong-answer-server doesn't exist, create it with 755 permissions
	os.MkdirAll(filepath.Dir(ConfigPath), 0755)

	// Write default config.yaml if not exists
	if _, err := os.Stat(ConfigPath); os.IsNotExist(err) {
		defaultContent := []byte(`port: "8080"
			answer_timer: "20"
			vote_timer: "180"
			`)
		if err := os.WriteFile(ConfigPath, defaultContent, 0644); err != nil {
			log.Fatal("Couldn't write default config file:", err)
		}
	}

	// Write default questions.json if not exists
	if _, err := os.Stat(QuestionsPath); os.IsNotExist(err) {
		defaultQuestions := []QuestionPair{ //In the future, server will pull questions from a public question API
			{Normal: "What's your favorite fruit?", Impostor: "What's your favorite cheese?"},
			{Normal: "What do you eat for breakfast?", Impostor: "What do you eat for dinner?"},
		}
		data, _ := json.MarshalIndent(defaultQuestions, "", "  ")
		if err := os.WriteFile(QuestionsPath, data, 0644); err != nil {
			log.Fatal("Couldn't write default questions file:", err)
		}
	}
}

func LoadConfig(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("could not read config file: %w", err)
	}

	return yaml.Unmarshal(data, &Config) //Unmarshals contents of config file into Config
}

func LoadQuestions(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("could not read questions file: %w", err)
	}

	return json.Unmarshal(data, &Questions) //unmarshals contents of questions file into Questions
}
