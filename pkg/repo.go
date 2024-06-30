package pkg

import (
	"context"
	"github.com/google/go-github/v62/github"
)

type Repo interface {
	GetCommitFileContent(ctx context.Context, filePath, sha string) (string, error)
	GetPendingReviewCommits() ([]*github.RepositoryCommit, error)
	GetPullRequestReviews(ctx context.Context) ([]*github.PullRequestReview, error)
	SubmitCodeReview(ctx context.Context, commit github.RepositoryCommit, comments []*github.DraftReviewComment) error
}
