package pkg

import (
	"context"
	"fmt"
	"github.com/google/go-github/v62/github"
	"os"
	"strings"
	"sync"
)

type Runner interface {
	RunCodeReview(ctx context.Context) error
}

type CodeReviewRunner struct {
	githubCli Repo
	config    *Config
	gptCli    Cli
}

func NewCodeReviewRunner(config *Config, repoCli Repo, gptCli Cli) (Runner, error) {
	if err := ValidateConfig(*config); err != nil {
		return nil, err
	}
	return &CodeReviewRunner{
		githubCli: repoCli,
		gptCli:    gptCli,
		config:    config,
	}, nil
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
		if !isFileSuffixAllowed(*file.Filename, c.config.ReviewSuffixes) {
			fmt.Fprintf(os.Stdout, "Skipping file: %s. The file does not have an allowed suffix (%v).\n", file, c.config.ReviewSuffixes)
			continue
		}
		pos := parseDiff(*file.Patch)
		cc, err := c.githubCli.GetCommitFileContent(ctx, *file.Filename, *commit.SHA)
		if err != nil {
			fmt.Println("error getting file content:", err)
			return nil, err
		}
		if c.config.MaxLineNum > 0 && len(strings.Split(cc, "\n")) > c.config.MaxLineNum {
			fmt.Fprintf(os.Stdout, "file[%s] line num is too large, skip\n", *file.Filename)
			continue
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

func (c *CodeReviewRunner) getReadyCommits(ctx context.Context) ([]*github.RepositoryCommit, error) {
	pendingCommits, err := c.githubCli.GetPendingReviewCommits()
	if err != nil {
		return nil, err
	}

	if c.config.MaxCommitNum > 0 && len(pendingCommits) > c.config.MaxCommitNum {
		fmt.Fprintf(os.Stdout, "no pending commits need to review\n")
		return nil, fmt.Errorf("no pending commits need to review")
	}

	var alreadyCommits []*github.RepositoryCommit

	for _, commit := range pendingCommits {
		if c.isHaveCodeReviewRecord(ctx, commit) {
			fmt.Fprintf(os.Stdout, "commit[%s]: %s  already reviewed\n", *commit.SHA, *commit.Commit.Message)
			continue
		}
		alreadyCommits = append(alreadyCommits, commit)
	}

	return alreadyCommits, nil

}

func (c *CodeReviewRunner) RunCodeReview(ctx context.Context) error {

	alreadyCommits, err := c.getReadyCommits(ctx)
	if err != nil {
		return err
	}

	if len(alreadyCommits) == 0 {
		fmt.Fprintf(os.Stdout, "No commits are ready for review.\n")
		return nil
	}

	concurrency := 1
	if c.config.MaxConcurrency > 0 {
		concurrency = c.config.MaxConcurrency
	}
	if concurrency > len(alreadyCommits) {
		concurrency = len(alreadyCommits)
	}

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, concurrency)

	for _, commit := range alreadyCommits {
		if c.isHaveCodeReviewRecord(ctx, commit) {
			fmt.Fprintf(os.Stdout, "commit[%s]: %s  already reviewed\n", *commit.SHA, *commit.Commit.Message)
			continue
		}
		if c.config.MaxFileNum > 0 && len(commit.Files) > c.config.MaxFileNum {
			fmt.Fprintf(os.Stdout, "commit[%s]: %s  file num is too large, skip\n", *commit.SHA, *commit.Commit.Message)
			continue
		}
		wg.Add(1)
		semaphore <- struct{}{} // Acquire semaphore
		go func(commit *github.RepositoryCommit) {
			defer wg.Done()
			defer func() {
				<-semaphore
			}()
			fmt.Fprintf(os.Stdout, "now reviewing commit[%s]: %s\n", *commit.SHA, *commit.Commit.Message)
			comments, err := c.reviewCommit(ctx, commit)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error reviewing commit, commit[%s]: %s, err :%s\n", *commit.SHA, *commit.Commit.Message, err)
				return
			}
			fmt.Fprintf(os.Stdout, "found %d comments in commit[%s]: %s \n", len(comments), *commit.SHA, *commit.Commit.Message)
			if comments == nil {
				return
			}
			// review end, submit the code review with this commit
			err = c.githubCli.SubmitCodeReview(ctx, *commit, comments)
			if err != nil {
				fmt.Fprintf(os.Stderr, "submit code review error, commit[%s]: %s, err: %s\n", *commit.SHA, *commit.Commit.Message, err)
				return
			}
			fmt.Fprintf(os.Stdout, "submit code review success, commit[%s]: %s, comment nums: %d\n", *commit.SHA, *commit.Commit.Message, len(comments))
		}(commit)
	}
	wg.Wait()
	return nil
}
