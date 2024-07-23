package pkg

import "fmt"

type Config struct {
	Owner          string   `yaml:"owner"`
	Repo           string   `yaml:"repo"`
	Pr             int      `yaml:"pr"`
	Token          string   `yaml:"token"`
	Prompt         string   `yaml:"prompt"`
	ApiKey         string   `yaml:"api_key"`
	MaxFileNum     int      `yaml:"max_file_num"`
	MaxLineNum     int      `yaml:"max_line_num"`
	MaxCommitNum   int      `yaml:"max_commit_num"`
	ReviewSuffixes []string `yaml:"review_suffixes"`
}

func ValidateConfig(config Config) error {

	if config.Pr <= 0 {
		return fmt.Errorf("config.pr is invailed")
	}

	if config.Owner == "" {
		return fmt.Errorf("config.owner is invailed")
	}

	if config.ApiKey == "" {
		return fmt.Errorf("config.api_key is invailed")
	}

	if config.Token == "" {
		return fmt.Errorf("config.token is invailed")
	}

	if config.Repo == "" {
		return fmt.Errorf("config.repo is invailed")
	}

	return nil
}
