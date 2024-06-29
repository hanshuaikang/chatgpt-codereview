package pkg

import "context"

type Cli interface {
	Send(ctx context.Context, content string) (string, error)
}
