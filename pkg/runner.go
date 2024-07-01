package pkg

import (
	"context"
	"fmt"
	"github.com/google/go-github/v62/github"
	"os"
)

type Runner interface {
	RunCodeReview(ctx context.Context) error
}

type CodeReviewRunner struct {
	githubCli Repo
	config    *Config
	gptCli    Cli
}

func NewCodeReviewRunner(config *Config, repoCli Repo, gptCli Cli) Runner {
	return &CodeReviewRunner{
		githubCli: repoCli,
		gptCli:    gptCli,
		config:    config,
	}
}

func (c *CodeReviewRunner) isInChanged(line int, pos map[int]int) bool {
	for start, end := range pos {
		if line >= start && line <= end {
			return true
		}
	}
	return false

}

// reviewCommit call chatgpt client and review a commit files
func (c *CodeReviewRunner) reviewCommit(ctx context.Context, commit *github.RepositoryCommit) ([]*github.DraftReviewComment, error) {

	var comments []*github.DraftReviewComment
	fmt.Fprintf(os.Stdout, "this commit has %d files need to review \n", len(commit.Files))
	for _, file := range commit.Files {
		pos := parseDiff(*file.Patch)
		cc, err := c.githubCli.GetCommitFileContent(ctx, *file.Filename, *commit.SHA)
		if err != nil {
			fmt.Println("error getting file content:", err)
			return nil, err
		}
		content := buildParam(c.config.Prompt, cc)
		reviewedReplyContent, err := c.gptCli.Send(ctx, content)
		if err != nil {
			fmt.Println("error making request:", err)
			return nil, err
		}
		gptComments := parseComments(reviewedReplyContent)
		for k, v := range gptComments {

			if !c.isInChanged(k, pos) {
				continue
			}
			comments = append(comments, &github.DraftReviewComment{
				Path: file.Filename,
				Line: &k,
				Body: &v},
			)
		}
	}

	return comments, nil
}

func (c *CodeReviewRunner) isHaveCodeReviewRecord(ctx context.Context, commit *github.RepositoryCommit) bool {
	reviews, err := c.githubCli.GetPullRequestReviews(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error getting reviews: %s\n", err)
		return false
	}
	for _, review := range reviews {
		if *review.CommitID == *commit.SHA {
			return true
		}
	}
	return false

}

func (c *CodeReviewRunner) RunCodeReview(ctx context.Context) error {

	pendingCommits, err := c.githubCli.GetPendingReviewCommits()
	if err != nil {
		return err
	}

	for _, commit := range pendingCommits {
		if c.isHaveCodeReviewRecord(ctx, commit) {
			fmt.Fprintf(os.Stdout, "commit[%s]: %s  already reviewed\n", *commit.SHA, *commit.Commit.Message)
			continue
		}
		fmt.Fprintf(os.Stdout, "now reviewing commit[%s]: %s\n", *commit.SHA, *commit.Commit.Message)
		comments, err := c.reviewCommit(ctx, commit)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error reviewing commit, commit[%s]: %s, err :%s\n", *commit.SHA, *commit.Commit.Message, err)
			return err
		}
		fmt.Fprintf(os.Stdout, "found %d comments in commit[%s]: %s \n", len(comments), *commit.SHA, *commit.Commit.Message)
		if comments == nil {
			continue
		}
		// review end , submit the code review with this commit
		err = c.githubCli.SubmitCodeReview(ctx, *commit, comments)

		if err != nil {
			fmt.Fprintf(os.Stderr, "submit code review error, commit[%s]: %s, err: %s\n", *commit.SHA, *commit.Commit.Message, err)
			return err
		}
		fmt.Fprintf(os.Stdout, "submit code review success, commit[%s]: %s, comment nums: %d\n", *commit.SHA, *commit.Commit.Message, len(comments))
	}

	return nil
}
