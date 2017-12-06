package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

var (
	repo                 string
	fromCommit, toCommit string
	configPath           string
)

func init() {
	defaultConfigPath := fmt.Sprintf("%s/.prlog.yaml", os.Getenv("HOME"))

	flag.StringVar(&repo, "repo", "", "'organization/name' of repository")
	flag.StringVar(&fromCommit, "from", "", "first commit of the changelog")
	flag.StringVar(&toCommit, "to", "", "last commit of the changelog")
	flag.StringVar(&configPath, "config", defaultConfigPath, "last commit of the changelog")

	viper.SetConfigType("yaml")

}

func main() {
	flag.Parse()

	viper.SetDefault("git-token", os.Getenv("PRLOG_GIT_TOKEN"))

	if configPath != "" {
		viper.SetConfigFile(configPath)
	}

	// Read configuration file
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Failed to read configuration: %v", err)
	}

	if repo == "" || fromCommit == "" || toCommit == "" {
		fmt.Println("Parameters 'repo', 'from' and 'to' are required:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	gitPRs, err := mergedPrs()
	if err != nil {
		fmt.Printf("Failed to list merged PRs from git history: %v\n", err)
		os.Exit(1)
	}

	ghPRs, err := githubPRs(repo)
	if err != nil {
		fmt.Printf("Failed to get PR details from GitHub: %s\n", err)
		os.Exit(1)
	}

	prlabels := viper.GetStringMapStringSlice("pr-labels")
	prGroups := make(map[string][]PullRequest)

	for _, gitPRs := range gitPRs {
		pr, err := pullRequestDesc(gitPRs, ghPRs)
		if err != nil {
			continue
		}

		assignToGroup(pr, prlabels, prGroups)
	}

	for grName, prs := range prGroups {
		printPRs(grName, prs)
	}
}

func printPRs(title string, prs []PullRequest) {
	if len(prs) != 0 {
		fmt.Printf("%s:\n", strings.Title(title))
		for _, pr := range prs {
			fmt.Printf(" - %s (#%d)\n", pr.Title, pr.Number)
		}
		fmt.Println()
	}
}

func pullRequestDesc(no int, all []PullRequest) (PullRequest, error) {
	for _, pr := range all {
		if pr.Number == no {
			return pr, nil
		}
	}
	return PullRequest{}, errors.Errorf("failed to find pull request #%d", no)
}

func assignToGroup(pr PullRequest, prlabels map[string][]string, prGroups map[string][]PullRequest) {
	assigned := false
	for _, lbl := range pr.LabelNames() {
		for grName, grLbls := range prlabels {
			for _, grLbl := range grLbls {
				if grLbl == lbl {
					prGroups[grName] = append(prGroups[grName], pr)
					assigned = true
				}
			}
		}
	}
	if !assigned {
		prGroups["other"] = append(prGroups["other"], pr)
	}
}
