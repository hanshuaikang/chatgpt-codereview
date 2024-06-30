package pkg

import "context"

type CodeReviewRunner interface {
	RunCodeReview(ctx context.Context) error
}
