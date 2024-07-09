package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

var httpsRemoteRe = regexp.MustCompile(`^https://([^/]+)/(.+?)(?:\.git)?$`)
var sshRemoteRe = regexp.MustCompile(`^git@([^:]+):(.+?)(?:\.git)?$`)

type ReleaseCreator interface {
	IsCorrectServer() bool
	CreateRelease(newVersion string, releaseNotes string) error
	Name() string
}

func getReleaseCreator(projectName string) (ReleaseCreator, error) {
	git := GitWrapper{}
	remoteURL, err := git.GetOriginRemoteURL()
	if err != nil {
		return nil, fmt.Errorf("error getting origin remote URL: %w", err)
	}

	serverURL := ""
	if strings.HasPrefix(remoteURL, "git@") {
		matches := sshRemoteRe.FindStringSubmatch(remoteURL)
		if len(matches) != 3 {
			return nil, fmt.Errorf("error parsing remote URL: %s", remoteURL)
		}

		serverURL = matches[1]
	} else {
		matches := httpsRemoteRe.FindStringSubmatch(remoteURL)
		if len(matches) != 3 {
			return nil, fmt.Errorf("error parsing remote URL: %s", remoteURL)
		}

		serverURL = matches[1]
	}

	// Currently only support GitLab, could support GitHub in the future.
	apiKey := os.Getenv("GITLAB_API_KEY")
	if apiKey != "" {
		creator, err := NewGitLabReleaseCreator(apiKey, serverURL, projectName)
		if err != nil {
			return nil, fmt.Errorf("error creating GitLab release creator: %w", err)
		}

		if creator.IsCorrectServer() {
			return creator, nil
		}
	}

	// Could also support GitHub here.
	return nil, fmt.Errorf("no supported release creator found")
}
