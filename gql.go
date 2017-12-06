package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

func githubPRs(repo string) ([]PullRequest, error) {
	flag.Parse()

	repoParts := strings.Split(repo, "/")
	orgName := repoParts[0]
	repoName := repoParts[1]

	changelogQuery := fmt.Sprintf(queryFmt, orgName, repoName)
	var changelog ChangelogResponse

	if err := queryGithubGQL(changelogQuery, &changelog); err != nil {
		return nil, errors.Wrap(err, "failed to query Github GQL")
	}

	return changelog.PullRequests(), nil
}

const queryFmt = `{
	"query": "query {
		repository(owner: \"%s\", name: \"%s\") {
			pullRequests(states: MERGED, last: 100) {
				nodes {
					number,
					title,
					labels(first: 10) {
						nodes {
							name
						}
					}
				}
			}
		}
	}"
}`

type PullRequestType int

const (
	UNKNOWN PullRequestType = iota
	FEATURE
	FIX
)

type PullRequest struct {
	Number int    `json:"number,omitempty"`
	Title  string `json:"title,omitempty"`
	Labels struct {
		Nodes []struct {
			Name string `json:"name,omitempty"`
		} `json:"nodes,omitempty"`
	} `json:"labels,omitempty"`
}

func (pr PullRequest) LabelNames() []string {
	var labels []string
	for _, lbl := range pr.Labels.Nodes {
		labels = append(labels, lbl.Name)
	}
	return labels
}

func (pr PullRequest) Type() PullRequestType {
	for _, lbl := range pr.LabelNames() {
		if lbl == "feature" {
			return FEATURE
		}
		if lbl == "fix" || lbl == "bug" {
			return FIX
		}
	}
	return UNKNOWN
}

type ChangelogResponse struct {
	Data struct {
		Repository struct {
			PullRequests struct {
				Nodes []PullRequest `json:"nodes,omitempty"`
			} `json:"pullRequests,omitempty"`
		} `json:"repository,omitempty"`
	} `json:"data,omitempty"`
}

func (cr ChangelogResponse) PullRequests() []PullRequest {
	return cr.Data.Repository.PullRequests.Nodes
}

func queryGithubGQL(query string, response interface{}) error {
	bearer := fmt.Sprintf("bearer %s", viper.GetString("git-token"))

	var queryBytes = []byte(toOneLine(query))
	req, _ := http.NewRequest("POST", "https://api.github.com/graphql", bytes.NewBuffer(queryBytes))
	req.Header.Set("Authorization", bearer)

	c := http.Client{}
	resp, err := c.Do(req)
	if err != nil {
		return errors.Wrap(err, "failed to query Github GQL")
	}

	if resp.StatusCode == http.StatusUnauthorized {
		return errors.New("unauthorized")
	}

	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return errors.Wrap(err, "failed to decode Github GQL response")
	}

	return nil
}
func toOneLine(q string) string {
	parts := strings.Split(q, "\n")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return strings.Join(parts, " ")
}
