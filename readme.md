# ChatGpt Code Review
[English Document](https://github.com/hanshuaikang/chatgpt-codereview/blob/main/readme_en.md)

ChatGpt Code Review 是一个用Golang 编写的 AI code Review 工具，它可以利用AI自动某个Github PR
的 Review ，并 comment 到对应的代码段下。效果如下:

![img.png](docs/imgs/img.png)

## Feature:
- PR 下的 commit 只会 Review 一次
- 基于全文件的Review 比基于 diff 的质量更高

## Usage

有两种方式使用 ChatGpt Code Review

### 第一种:

作为 command 使用。

1. 创建一个 config.yaml 文件:

```yaml
token: "github token"
owner: "repo owner"
repo: "repo name"
api_key: "openai key"
pr: pr Id
```

```bash
go install github.com/hanshuaikang/chatgpt-codereview@latest

chatgpt-codereview run ./config.yaml
```

2. 引入代码使用

```bash
go get github.com/hanshuaikang/chatgpt-codereview@latest
```

main.go
```bash
package main

import (
	"context"
	"github.com/hanshuaikang/chatgpt-codereview/pkg"
	"github.com/hanshuaikang/chatgpt-codereview/pkg/chatgpt"
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
	}
	gpt := chatgpt.NewChatGpt(&config)

	ctx := context.Background()
	err := gpt.RunCodeReview(ctx)
	if err != nil {
		os.Exit(1)
	}
}
```


