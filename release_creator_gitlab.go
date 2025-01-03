package main

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/xanzy/go-gitlab"
)

type GitLabReleaseCreator struct {
	gitlabClient *gitlab.Client
	projectName  string
}

func NewGitLabReleaseCreator(conf *Config, baseURL string, projectName string) (*GitLabReleaseCreator, error) {
	gitlabClient, err := gitlab.NewClient(conf.GitlabAPIKey, gitlab.WithBaseURL(fmt.Sprintf("https://%s", baseURL)))
	if err != nil {
		return nil, fmt.Errorf("error creating GitLab client: %w", err)
	}

	return &GitLabReleaseCreator{
		gitlabClient: gitlabClient,
		projectName:  projectName,
	}, nil
}

// IsCorrectServer returns true if the specified base URL actually points to a GitLab server.
func (g *GitLabReleaseCreator) IsCorrectServer() bool {
	if _, _, err := g.gitlabClient.Version.GetVersion(); errors.Is(err, gitlab.ErrNotFound) {
		return false
	}

	// 401 error is fine - we might not have specified the API key yet.
	return true
}

func (g *GitLabReleaseCreator) CreateRelease(newVersion string, releaseNotes string) (*url.URL, error) {
	projectID, err := g.getProjectID()
	if err != nil {
		return nil, fmt.Errorf("error getting project ID: %w", err)
	}

	opts := gitlab.CreateReleaseOptions{
		Name:        gitlab.Ptr(fmt.Sprintf("%s %s", g.projectName, newVersion)),
		TagName:     &newVersion,
		Description: &releaseNotes,
	}

	release, _, err := g.gitlabClient.Releases.CreateRelease(projectID, &opts)
	if err != nil {
		return nil, fmt.Errorf("error creating release: %w", err)
	}

	releaseURL, err := url.Parse(release.Links.Self)
	if err != nil {
		return nil, fmt.Errorf("error parsing release URL: %w", err)
	}

	return releaseURL, nil
}

func (g *GitLabReleaseCreator) Name() string {
	return "GitLab"
}

func (g *GitLabReleaseCreator) getProjectID() (int, error) {
	git := GitWrapper{}
	remoteURL, err := git.GetOriginRemoteURL()
	if err != nil {
		return 0, fmt.Errorf("error getting origin remote URL: %w", err)
	}

	projectPath := ""
	if strings.HasPrefix(remoteURL, "git@") {
		matches := sshRemoteRe.FindStringSubmatch(remoteURL)
		if len(matches) != 3 {
			return 0, fmt.Errorf("error parsing remote URL: %s", remoteURL)
		}

		projectPath = matches[2]
	} else {
		matches := httpsRemoteRe.FindStringSubmatch(remoteURL)
		if len(matches) != 3 {
			return 0, fmt.Errorf("error parsing remote URL: %s", remoteURL)
		}

		projectPath = matches[2]
	}

	// You can actually just use the project path for all API calls, but getting the proper project ID
	// allows us to verify that the project exists.
	project, _, err := g.gitlabClient.Projects.GetProject(projectPath, nil)
	if err != nil {
		return 0, fmt.Errorf("error getting project: %w", err)
	}

	return project.ID, nil
}
