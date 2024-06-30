package chatgpt

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

func parseDiff(diff string) (int, int) {
	lines := strings.Split(diff, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "@@") {
			// 解析 @@ -106,46 +106,6 @@ 这样的行
			parts := strings.Split(line, " ")
			if len(parts) < 4 {
				continue
			}

			// 解析 +106,6 部分
			addedInfo := parts[2]
			addedParts := strings.Split(addedInfo[1:], ",")
			if len(addedParts) != 2 {
				continue
			}

			startLine, err1 := strconv.Atoi(addedParts[0])
			numLines, err2 := strconv.Atoi(addedParts[1])
			if err1 != nil || err2 != nil {
				continue
			}

			endLine := startLine + numLines - 1
			return startLine, endLine
		}
	}

	return 0, 0
}

func parseComments(input string) map[int]string {
	scanner := bufio.NewScanner(strings.NewReader(input))
	parsedComments := map[int]string{}

	// 正则表达式匹配行号和评论内容
	re := regexp.MustCompile(`\[(\d+)\] (.+)`)

	for scanner.Scan() {
		line := scanner.Text()
		matches := re.FindStringSubmatch(line)
		if matches == nil {
			continue
		}

		lineNumber := matches[1]
		codeContent := matches[2]
		l, _ := strconv.Atoi(lineNumber)
		parsedComments[l] = codeContent
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %s\n", err)
		return nil
	}

	return parsedComments
}
