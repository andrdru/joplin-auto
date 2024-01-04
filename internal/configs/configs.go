package configs

import (
	"fmt"
	"os"
	"strings"

	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
)

type (
	// Config main config
	Config struct {
		IsDebug  bool   `yaml:"is_debug"`
		AppID    string `yaml:"app_id"`
		NoteID   string `yaml:"note_id"`
		ParentID string `yaml:"parent_id"`
		S3       S3     `yaml:"s3"`
	}

	S3 struct {
		Host   string `yaml:"host"`
		Key    string `yaml:"key"`
		Secret string `yaml:"secret"`
		Bucket string `yaml:"bucket"`
		Region string `yaml:"region"`
	}
)

// NewConfig read config from file
func NewConfig(path string) (config Config, err error) {
	var bytes []byte
	bytes, err = os.ReadFile(path)
	if err != nil {
		return config, fmt.Errorf("could not read file: %w", err)
	}

	if err = yaml.Unmarshal(bytes, &config); err != nil {
		return config, fmt.Errorf("could not unmarshal config: %w", err)
	}

	if config.AppID == "" {
		config.AppID = GenerateID()
	}

	return config, nil
}

func GenerateID() string {
	return strings.ReplaceAll(uuid.New().String(), "-", "")
}
