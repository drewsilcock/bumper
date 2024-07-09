package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type GitWrapper struct{}

func (g *GitWrapper) GetCurrentBranch() (string, error) {
	getCurrentBranch := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	output, err := getCurrentBranch.Output()
	if err != nil {
		return "", fmt.Errorf("error getting current branch: %w", err)
	}

	currentBranch := strings.TrimSpace(string(output))
	return currentBranch, nil
}

func (g *GitWrapper) GetLatestTag() (string, error) {
	getLatestTag := exec.Command("git", "describe", "--tags", "--abbrev=0")
	output, err := getLatestTag.Output()
	if err != nil {
		return "", fmt.Errorf("error getting latest tag: %w", err)
	}

	latestTag := strings.TrimSpace(string(output))
	return latestTag, nil
}

func (g *GitWrapper) CreateBranch(branchName string) error {
	createBranch := exec.Command("git", "checkout", "-b", branchName)
	if err := createBranch.Run(); err != nil {
		return fmt.Errorf("error creating branch: %w", err)
	}

	return nil
}

func (g *GitWrapper) CheckoutBranch(branchName string) error {
	checkoutBranch := exec.Command("git", "checkout", branchName)
	if err := checkoutBranch.Run(); err != nil {
		return fmt.Errorf("error checking out branch: %w", err)
	}

	return nil
}

func (g *GitWrapper) MergeBranch(branchName string) error {
	mergeBranch := exec.Command("git", "merge", branchName, "--no-ff")
	if err := mergeBranch.Run(); err != nil {
		return fmt.Errorf("error merging branch: %w", err)
	}

	return nil
}

func (g *GitWrapper) Tag(tagName string) error {
	tag := exec.Command("git", "tag", tagName)
	if err := tag.Run(); err != nil {
		return fmt.Errorf("error tagging: %w", err)
	}

	return nil
}

func (g *GitWrapper) Push() error {
	push := exec.Command("git", "push")
	if err := push.Run(); err != nil {
		return fmt.Errorf("error pushing: %w", err)
	}

	return nil
}

func (g *GitWrapper) PushTags() error {
	pushTags := exec.Command("git", "push", "--tags")
	if err := pushTags.Run(); err != nil {
		return fmt.Errorf("error pushing tags: %w", err)
	}

	return nil
}

func (g *GitWrapper) DeleteBranch(branchName string) error {
	deleteBranch := exec.Command("git", "branch", "-d", branchName)
	if err := deleteBranch.Run(); err != nil {
		return fmt.Errorf("error deleting branch: %w", err)
	}

	return nil
}

func (g *GitWrapper) Commit(message string) error {
	commit := exec.Command("git", "commit", "-am", message)
	if err := commit.Run(); err != nil {
		return fmt.Errorf("error committing: %w", err)
	}

	return nil
}

func (g *GitWrapper) Add(path string) error {
	add := exec.Command("git", "add", path)
	if err := add.Run(); err != nil {
		return fmt.Errorf("error adding: %w", err)
	}

	return nil
}

func (g *GitWrapper) HasUncommittedChanges() (bool, error) {
	status := exec.Command("git", "status", "--porcelain")
	output, err := status.Output()
	if err != nil {
		return false, fmt.Errorf("error getting git status: %w", err)
	}

	return strings.TrimSpace(string(output)) != "", nil
}

func (g *GitWrapper) RunDiff(staged bool) {
	stagedArg := ""
	if staged {
		stagedArg = "--staged"
	}
	status := exec.Command("git", "diff", stagedArg)
	status.Stdout = os.Stdout
	status.Stderr = os.Stderr
	_ = status.Run()
}

func (g *GitWrapper) GetOriginRemoteURL() (string, error) {
	originRemoteCmd := exec.Command("git", "remote", "get-url", "origin")
	output, err := originRemoteCmd.Output()
	if err != nil {
		return "", fmt.Errorf("error getting origin remote URL: %v\n", err)
	}

	remoteURL := strings.TrimSpace(string(output))
	return remoteURL, nil
}
