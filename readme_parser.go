package main

import (
	"fmt"
	"io"
	"os"
	"path"
	"regexp"
)

var readmeNameRe = regexp.MustCompile(`^# (.+?)\n`)

type ReadmeParser struct {
	projectPath string
}

func NewReadmeParser(projectPath string) *ReadmeParser {
	return &ReadmeParser{projectPath: projectPath}
}

func (r *ReadmeParser) GetProjectName() (string, error) {
	// We assume that the h1 of the README contains the project name.
	readmeFile, err := os.Open(path.Join(r.projectPath, "README.md"))
	if err != nil {
		return "", fmt.Errorf("error opening README: %w", err)
	}
	defer readmeFile.Close()

	readmeBytes, err := io.ReadAll(readmeFile)
	if err != nil {
		return "", fmt.Errorf("error reading README: %w", err)
	}

	readmeContents := string(readmeBytes)

	matches := readmeNameRe.FindStringSubmatch(readmeContents)
	if len(matches) != 2 {
		return "", fmt.Errorf("project name not found in README")
	}

	return matches[1], nil
}
