package chatgpt

import (
	"bufio"
	"context"
	"fmt"
	"github.com/google/go-github/v62/github"
	"github.com/hanshuaikang/chatgpt-codereview/pkg"
	githubCli "github.com/hanshuaikang/chatgpt-codereview/pkg/github"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type ChatGpt struct {
	githubCli *githubCli.Client
	config    *pkg.Config
	gptCli    pkg.Cli
}

func NewChatGpt(config *pkg.Config) *ChatGpt {
	return &ChatGpt{
		githubCli: githubCli.NewGithubCli(config.Token, config.Owner, config.Repo, config.Pr),
		gptCli:    NewGptClient(config),
		config:    config,
	}
}

func (c *ChatGpt) parseAndApplyDiff(diffContent string) (string, error) {
	scanner := bufio.NewScanner(strings.NewReader(diffContent))
	hunkRegex := regexp.MustCompile(`^@@ -(\d+),(\d+) \+(\d+),(\d+) @@`)

	var modifiedContent []string
	lineNumber := 0

	for scanner.Scan() {
		line := scanner.Text()

		// 检查是否为 hunk 头部
		if matches := hunkRegex.FindStringSubmatch(line); matches != nil {
			startLine, _ := strconv.Atoi(matches[3])
			lineNumber = startLine - 1
			continue
		}

		// 处理添加的行
		if strings.HasPrefix(line, "+") {
			lineNumber++
			modifiedContent = append(modifiedContent, fmt.Sprintf("%d %s", lineNumber, line[1:]))
		}

		// 处理未修改的行
		if strings.HasPrefix(line, " ") {
			lineNumber++
			modifiedContent = append(modifiedContent, fmt.Sprintf("%d %s", lineNumber, line[1:]))
		}
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return strings.Join(modifiedContent, "\n"), nil
}

func (c *ChatGpt) reviewCommit(ctx context.Context, commit *github.RepositoryCommit) ([]*github.DraftReviewComment, error) {

	var comments []*github.DraftReviewComment
	fmt.Fprintf(os.Stdout, "this commit has %d files need to review \n", len(commit.Files))
	for _, file := range commit.Files {
		diffStart, diffEnd := parseDiff(*file.Patch)
		cc, err := c.githubCli.GetCommitFileContent(ctx, *file.Filename, *commit.SHA)
		if err != nil {
			fmt.Println("Error getting file content:", err)
			return nil, err
		}
		reviewedReplyContent, err := c.gptCli.Send(ctx, cc)
		if err != nil {
			fmt.Println("Error making request:", err)
			return nil, err
		}
		gptComments := parseComments(reviewedReplyContent)
		for k, v := range gptComments {
			if k < diffStart || k > diffEnd {
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

func (c *ChatGpt) isHaveCodeReviewRecord(ctx context.Context, commit *github.RepositoryCommit) bool {
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

func (c *ChatGpt) RunCodeReview(ctx context.Context) error {

	pendingCommits, err := c.githubCli.GetPendingReviewCommits()
	if err != nil {
		return err
	}

	for _, commit := range pendingCommits {
		if c.isHaveCodeReviewRecord(ctx, commit) {
			fmt.Fprintf(os.Stdout, "commit[%s]: %s  already reviewed\n", *commit.SHA, *commit.Commit.Message)
			continue
		}
		fmt.Fprintf(os.Stdout, "now reviewing commit[%s]: %s", *commit.SHA, *commit.Commit.Message)
		comments, err := c.reviewCommit(ctx, commit)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error reviewing commit, commit[%s]: %s, err :%s", *commit.SHA, *commit.Commit.Message, err)
			return err
		}
		fmt.Fprintf(os.Stdout, "found %d comments in commit[%s]: %s \n", len(comments), *commit.SHA, *commit.Commit.Message)
		if comments == nil {
			continue
		}
		err = c.githubCli.SubmitCodeReview(ctx, *commit, comments)

		if err != nil {
			fmt.Fprintf(os.Stderr, "submit code review error, commit[%s]: %s, err: %s", *commit.SHA, *commit.Commit.Message, err)
			return err
		}
		fmt.Fprintf(os.Stdout, "submit code review success, commit[%s]: %s, comment nums: %d", *commit.SHA, *commit.Commit.Message, len(comments))
	}

	return nil
}
