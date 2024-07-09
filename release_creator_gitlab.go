package main

import (
	"fmt"
	"github.com/xanzy/go-gitlab"
	"strings"
)

type GitLabReleaseCreator struct {
	gitlabClient *gitlab.Client
	projectName  string
}

func NewGitLabReleaseCreator(apiKey string, baseURL string, projectName string) (*GitLabReleaseCreator, error) {
	gitlabClient, err := gitlab.NewClient(apiKey, gitlab.WithBaseURL(fmt.Sprintf("https://%s", baseURL)))
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
	if _, _, err := g.gitlabClient.Version.GetVersion(); err != nil {
		return false
	}

	return true
}

func (g *GitLabReleaseCreator) CreateRelease(newVersion string, releaseNotes string) error {
	projectID, err := g.getProjectID()
	if err != nil {
		return fmt.Errorf("error getting project ID: %w", err)
	}

	opts := gitlab.CreateReleaseOptions{
		Name:        gitlab.Ptr(fmt.Sprintf("%s %s", g.projectName, newVersion)),
		TagName:     &newVersion,
		Description: &releaseNotes,
	}

	if _, _, err := g.gitlabClient.Releases.CreateRelease(projectID, &opts); err != nil {
		return fmt.Errorf("error creating release: %w", err)
	}

	return nil
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
