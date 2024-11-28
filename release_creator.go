package main

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/manifoldco/promptui"
)

var httpsRemoteRe = regexp.MustCompile(`^https://([^/]+)/(.+?)(?:\.git)?$`)
var sshRemoteRe = regexp.MustCompile(`^git@([^:]+):(.+?)(?:\.git)?$`)

type ReleaseCreator interface {
	IsCorrectServer() bool
	CreateRelease(newVersion string, releaseNotes string) (*url.URL, error)
	Name() string
}

func getReleaseCreator(projectName string, conf *Config) (ReleaseCreator, error) {
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

	creator, err := NewGitLabReleaseCreator(conf, serverURL, projectName)
	if err != nil {
		return nil, fmt.Errorf("error creating GitLab release creator: %w", err)
	}

	if creator.IsCorrectServer() {
		if conf.GitlabAPIKey == "" {
			prompt := promptui.Prompt{
				Label:       "Please specify a GitLab API key with 'api' permission",
				HideEntered: true,
			}

			result, err := prompt.Run()
			if err != nil {
				return nil, fmt.Errorf("error prompting for GitLab API key: %w", err)
			}

			conf.GitlabAPIKey = result
			if err := conf.Write(); err != nil {
				return nil, fmt.Errorf("error saving configuration: %w", err)
			}
		}

		return creator, nil
	}

	// Could also support GitHub here.
	return nil, fmt.Errorf("no supported release creator found")
}
