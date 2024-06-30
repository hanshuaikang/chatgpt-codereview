package cmd

import (
	"context"
	"fmt"
	"github.com/hanshuaikang/chatgpt-codereview/pkg"
	"github.com/hanshuaikang/chatgpt-codereview/pkg/chatgpt"
	"github.com/hanshuaikang/chatgpt-codereview/pkg/github"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"os"
)

const prompt = `
角色：你是一个非常超级高级的Golang工程师。

任务：请帮我review以下代码，重点审查是否存在非常严重的bug或代码逻辑错误。
只有在遇到[代码漏洞，逻辑漏洞，拼写错误，性能问题]等关键问题时才提出修改建议。请忽略补充日志、注释、优化错误，缺少结构体定义信息等常规性建议。

要求：

审查标准：严格遵循以下几点：
只提出关键性问题包括，代码漏洞、逻辑漏洞、拼写错误、性能问题。
建议格式：所有建议应按照“[行号] 建议内容”的格式提出，以便清晰地识别问题所在的具体位置。
返回内容限制：只返回有问题的行，不要返回任何标记为“无需修改”的行。

`

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "run subcommand start review the code",
	Run: func(cmd *cobra.Command, args []string) {
		configFilePath := cmd.Flag("config").Value.String()
		config, err := parseConfig(configFilePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "parse config file failed: %s\n", err)
			os.Exit(1)
		}
		if config.Prompt == "" {
			config.Prompt = prompt
		}
		defaultGptCli := chatgpt.NewGptClient(config)
		githubCli := github.NewGithubCli(config.Token, config.Owner, config.Repo, config.Pr)
		gpt := chatgpt.NewChatGpt(&config, githubCli, defaultGptCli)

		ctx := context.Background()
		err = gpt.RunCodeReview(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "run code review failed: %s\n", err)
			os.Exit(1)
		}
	},
}

func parseConfig(path string) (pkg.Config, error) {
	// 读取 YAML 文件
	file, err := os.Open(path)
	if err != nil {
		return pkg.Config{}, err
	}
	defer file.Close()

	var config pkg.Config
	decoder := yaml.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		return pkg.Config{}, err
	}

	if config.Pr <= 0 {
		return pkg.Config{}, fmt.Errorf("config.pr is invailed")
	}

	if config.Owner == "" {
		return pkg.Config{}, fmt.Errorf("config.owner is invailed")
	}

	if config.ApiKey == "" {
		return pkg.Config{}, fmt.Errorf("config.api_key is invailed")
	}

	if config.Token == "" {
		return pkg.Config{}, fmt.Errorf("config.token is invailed")
	}

	if config.Repo == "" {
		return pkg.Config{}, fmt.Errorf("config.repo is invailed")
	}

	return config, nil

}

func init() {
	rootCmd.AddCommand(runCmd)
}
