package main

import (
	"bytes"
	"fmt"
	"github.com/spf13/viper"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

func mergedPrs() ([]int, error) {
	commitArg := fmt.Sprintf("%s..%s", fromCommit, toCommit)

	var cmd *exec.Cmd

	if local := viper.GetBool("local-only"); local {
		cmd = exec.Command("git", "log", "--oneline", commitArg)
	} else {
		gitDirArg := fmt.Sprintf("--git-dir=%s/src/github.com/%s/.git", os.Getenv("GOPATH"), repo)
		cmd = exec.Command("git", gitDirArg, "log", "--oneline", commitArg)
	}

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return nil, errors.Wrap(err, "failed to list local git history")
	}
	gitLog := string(out.Bytes())
	commits := strings.Split(gitLog, "\n")

	var re = regexp.MustCompile(`#\d+`)

	var prs []int
	for _, match := range re.FindAllString(gitLog, -1) {
		prNoStr := strings.Trim(match, "#")
		prNo, err := strconv.Atoi(prNoStr)
		if err != nil {
			continue
		}
		if findRevertCommit(commits, match) {
			continue
		}
		prs = append(prs, prNo)
	}
	return prs, nil
}

func findRevertCommit(commits []string, prHash string) bool {
	for _, c := range commits {
		if hasRevertPrefix(c) && strings.Contains(c, prHash) {
			return true
		}
	}
	return false
}

// hasRevertPrefix checks if the commit message (which starts at the 8th character)
/// starts with 'Revert' which may be an indication that it is a PR revert commit.
func hasRevertPrefix(msg string) bool {
	if len(msg) < 14 {
		return false
	}

	return msg[8:14] == "Revert"
}
