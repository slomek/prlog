package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

func mergedPrs() ([]int, error) {
	gitDirArg := fmt.Sprintf("--git-dir=%s/src/github.com/%s/.git", os.Getenv("GOPATH"), repo)
	commitArg := fmt.Sprintf("%s..%s", fromCommit, toCommit)
	cmd := exec.Command("git", gitDirArg, "log", "--oneline", commitArg)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return nil, errors.Wrap(err, "failed to list local git history")
	}
	gitLog := string(out.Bytes())

	var re = regexp.MustCompile(`#\d+`)

	var prs []int
	for _, match := range re.FindAllString(gitLog, -1) {
		prNoStr := strings.Trim(match, "#")
		prNo, err := strconv.Atoi(prNoStr)
		if err != nil {
			continue
		}
		prs = append(prs, prNo)
	}
	return prs, nil
}
