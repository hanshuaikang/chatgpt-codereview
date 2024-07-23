<h1 align="center"> 🧐 ChatGpt Code Review</h1>
<p align="center">
    <em>ChatGpt Code Review is an AI code review tool written in Golang. It uses AI to automatically review a specific GitHub PR and comments on the relevant code sections. The effect is as follows:
</em>
</p>

![img.png](docs/imgs/img.png)

## Features:
- Commits under a PR will be reviewed only once.
- Review based on the entire file is of higher quality compared to review based on diffs.

## Usage

### Configuration Description

- `max_file_num`: If the number of files to be reviewed in the current commit exceeds this value, the commit review will be skipped. If set to 0, there is no limit.
- `max_line_num`: If the number of lines in a file to be reviewed exceeds this value, the file review will be skipped. If set to 0, there is no limit.
- `max_commit_num`: If the number of commits in a pull request to be reviewed exceeds this value, the PR review will be skipped. If set to 0, there is no limit.
- `review_suffixes`: When `review_suffixes` is specified, only files with the specified suffixes will be reviewed. If set to an empty array (`[]`), there is no restriction.
- `max_concurrency`: the maximum concurrent number of review at the same time, 1 when it is not set

There are two ways to use ChatGpt Code Review:

### Method 1:

Use as a command.

Create a config.yaml file:
```yaml
token: "github token"
owner: "repo owner"
repo: "repo name"
api_key: "openai key"
pr: pr Id
max_file_num: 10
max_line_num: 1000
max_commit_num: 100
max_concurrency: 3
review_suffixes:
  - ".go"
```

Install and run the tool:
```yaml
go install github.com/hanshuaikang/chatgpt-codereview@latest
chatgpt-codereview run ./config.yaml
```

### Method 2:
Import the code and use it in your project.

Get the package:
```bash
go get github.com/hanshuaikang/chatgpt-codereview@latest
```

Use it in your project:

```golang
package main

import (
	"context"
	"github.com/hanshuaikang/chatgpt-codereview/pkg"
	"github.com/hanshuaikang/chatgpt-codereview/pkg/chatgpt"
	"github.com/hanshuaikang/chatgpt-codereview/pkg/github"
	"os"
)

func main() {
	config := pkg.Config{
		Owner:  "",
		Repo:   "",
		Pr:     0,
		Prompt: "",
		ApiKey: "",
		Token:  "",
		MaxFileNum: 10,
		MaxLineNum: 1000,
		MaxCommitNum: 100,
		MaxConcurrency: 3,
		ReviewSuffixes: []string{".go"},
	}

	defaultGptCli := chatgpt.NewGptClient(config)
	githubCli := github.NewGithubCli(config.Token, config.Owner, config.Repo, config.Pr)
	runner, err := pkg.NewCodeReviewRunner(&config, githubCli, defaultGptCli)
	if err != nil {
		os.Exit(1)
	}
	ctx := context.Background()
	err = gpt.RunCodeReview(ctx)
	err = runner.RunCodeReview(ctx)
	if err != nil {
		os.Exit(1)
	}
}

```