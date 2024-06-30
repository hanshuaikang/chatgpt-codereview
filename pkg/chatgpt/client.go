package chatgpt

import (
	"context"
	"github.com/hanshuaikang/chatgpt-codereview/pkg"
	"github.com/sashabaranov/go-openai"
)

type GptClient struct {
	config pkg.Config
}

func (g GptClient) Send(ctx context.Context, content string) (string, error) {
	// call open api
	client := openai.NewClient(g.config.ApiKey)
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: content,
				},
			},
		},
	)

	if err != nil {
		return "", err
	}

	if len(resp.Choices) == 0 {
		return "", nil
	}

	return resp.Choices[0].Message.Content, nil
}

func NewGptClient(config pkg.Config) pkg.Cli {
	return &GptClient{
		config: config,
	}
}
