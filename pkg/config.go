package pkg

type Config struct {
	Owner  string `yaml:"owner"`
	Repo   string `yaml:"repo"`
	Pr     int    `yaml:"pr"`
	Token  string `yaml:"token"`
	Prompt string `yaml:"prompt"`
	ApiKey string `yaml:"api_key"`
}
