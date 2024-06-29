package chatgpt

import (
	"context"
	"fmt"
	"github.com/hanshuaikang/chatgpt-codereview/pkg"
	"github.com/sashabaranov/go-openai"
)

type GptClient struct {
	config *pkg.Config
}

func (g GptClient) buildParam(content string) string {
	prompt := "%s. You must return it in this format, like [25] if err ! = nil { . instead of [Line 25] if err ! = nil \n %s"
	return fmt.Sprintf(g.config.Prompt, prompt, content)
}

func (g GptClient) Send(ctx context.Context, content string) (string, error) {
	client := openai.NewClient(g.config.ApiKey)
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: g.buildParam(content),
				},
			},
		},
	)

	if err != nil {
		return "", err
	}

	if len(resp.Choices) == 0 {
		return "", err
	}

	return resp.Choices[0].Message.Content, nil
}

func NewGptClient(config *pkg.Config) pkg.Cli {
	return &GptClient{
		config: config,
	}
}
