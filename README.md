# prlog

The idea behind prlog is to join information stored in `git` history and GitHub's pull requests in order to create a changelog between two commits. The pull requests can be grouped by their labels to make the changelog even more readable to the users.

## Installation

In order to install the application, you need to `go get` it.

## Usage

### Configuration

First, you need to create a configuration YAML file (by default it is searched for in `${HOME}/.prlog.yaml`):

    git-token: YOUR-GIT-TOKEN

    pr-labels:
        features: 
            - Enhancement
        fixes:
            - fix
            - Bug
        internal:
            - refactor
            - infra

The `pr-labels` property works as follows: if a PR has a label _Enhancement_, it will be listed under a group called _features_ in the changelog. The _fixes_ group will contain all PRs having labels _fix_ or _bug_. Any PRs that are not assigned to any of the defined groups will land in _other_ section.

### Usage

In order to print a changelog, you need to have a git history of the repository fetched on your local machine. Then, all you need to do is execute the app:

    $ prlog -repo gobuffalo/buffalo -to caff091 -from ea3de19 -config prlog.yaml
    Other:
    - Fix Issue #751 - unknown flag: --use-model (#778)
    - Running a single test fixes #769 (#772)

    Fixes:
    - Fix skip-yarn flag description (#776)
    - Fix test render packr config (#770)
    - Fix #767: CSRF middleware should accept wildcard mimetype (#768)

    Features:
    - Implement #723 - Localized views (#771)

As you can see, the PRs have been grouped into sections.
