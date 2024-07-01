package github

import (
	"context"
	"fmt"
	"github.com/google/go-github/v62/github"
	"github.com/hanshuaikang/chatgpt-codereview/pkg"
	"os"
	"strings"
)

type Client struct {
	client *github.Client
	owner  string
	repo   string
	prId   int
}

func NewGithubCli(token string, owner string, repo string, prId int) pkg.Repo {
	return &Client{
		client: github.NewClient(nil).WithAuthToken(token),
		owner:  owner,
		repo:   repo,
		prId:   prId,
	}
}

func (c *Client) getCommits() ([]*github.RepositoryCommit, error) {
	ctx := context.Background()
	commits, _, err := c.client.PullRequests.ListCommits(ctx, c.owner, c.repo, c.prId, nil)
	if err != nil {
		return nil, err
	}
	return commits, nil
}

func (c *Client) GetCommitFileContent(ctx context.Context, filePath, sha string) (string, error) {
	fileContent, _, _, err := c.client.Repositories.GetContents(ctx, c.owner, c.repo, filePath, &github.RepositoryContentGetOptions{Ref: sha})
	if err != nil {
		return "", err
	}

	content, err := fileContent.GetContent()
	if err != nil {
		return "", err

	}
	var contentWithLine []string
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		contentWithLine = append(contentWithLine, fmt.Sprintf("%d %s", i+1, line))
	}

	return strings.Join(contentWithLine, "\n"), nil
}

func (c *Client) getCommit(owner, repo, sha string) (*github.RepositoryCommit, error) {
	ctx := context.Background()
	commit, _, err := c.client.Repositories.GetCommit(ctx, owner, repo, sha, nil)
	if err != nil {
		return nil, err
	}
	return commit, nil
}

func (c *Client) GetPendingReviewCommits() ([]*github.RepositoryCommit, error) {
	commits, err := c.getCommits()
	if err != nil {
		return nil, err
	}
	var pendingReviewCommits []*github.RepositoryCommit
	for _, commit := range commits {
		if strings.HasPrefix(*commit.Commit.Message, "Merge pull request") {
			continue
		}
		commitInfo, err := c.getCommit(c.owner, c.repo, *commit.SHA)
		if err != nil {
			return nil, err
		}
		pendingReviewCommits = append(pendingReviewCommits, commitInfo)
	}
	return pendingReviewCommits, nil
}

func (c *Client) GetPullRequestReviews(ctx context.Context) ([]*github.PullRequestReview, error) {
	reviews, _, err := c.client.PullRequests.ListReviews(ctx, c.owner, c.repo, c.prId, &github.ListOptions{})
	if err != nil {
		return nil, err
	}
	return reviews, nil
}

func (c *Client) SubmitCodeReview(ctx context.Context, commit github.RepositoryCommit, comments []*github.DraftReviewComment) error {

	review, _, err := c.client.PullRequests.CreateReview(ctx, c.owner, c.repo, c.prId, &github.PullRequestReviewRequest{
		CommitID: commit.SHA,
		Comments: comments,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating review: %s\n", err)
		return err
	}

	_, _, err = c.client.PullRequests.SubmitReview(ctx, c.owner, c.repo, c.prId, *review.ID, &github.PullRequestReviewRequest{
		Event: github.String("COMMENT"),
	})

	return err
}
