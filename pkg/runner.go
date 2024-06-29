package pkg

import "context"

type CodeReviewRunner interface {
	RunCodeReview(ctx context.Context, config Config, content string) (map[int]string, error)
}
