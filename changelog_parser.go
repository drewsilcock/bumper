package main

import (
	"fmt"
	"github.com/dustin/go-humanize"
	"io"
	"os"
	"path"
	"regexp"
	"time"
)

var unreleasedHeaderRe = regexp.MustCompile(`(?m)^## (Unreleased|Development)$`)
var unreleasedSectionRe = regexp.MustCompile(`(?ms)^## (Unreleased|Development)\n\n(.+?)(?:\n## |\z)`)

type ChangelogUpdater struct {
	filePath string
}

func NewChangelogUpdater(projectPath string) *ChangelogUpdater {
	return &ChangelogUpdater{filePath: path.Join(projectPath, "CHANGELOG.md")}
}

func (c *ChangelogUpdater) GetUnreleasedNotes() (string, error) {
	file, err := os.Open(c.filePath)
	if err != nil {
		return "", fmt.Errorf("error opening CHANGELOG.md: %w", err)
	}
	defer file.Close()

	changelogBytes, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("error reading CHANGELOG.md: %w", err)
	}

	changelogContents := string(changelogBytes)

	// `matches[0]` contains full matching string, then rest of slice contains the capture groups in order.
	matches := unreleasedSectionRe.FindStringSubmatch(changelogContents)
	if len(matches) < 3 {
		return "", fmt.Errorf("unreleased section not found in CHANGELOG.md")
	}

	return matches[2], nil
}

func (c *ChangelogUpdater) Update(newVersion string) error {
	file, err := os.Open(c.filePath)
	if err != nil {
		return fmt.Errorf("error opening CHANGELOG.md: %w", err)
	}
	defer file.Close()

	changelogBytes, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("error reading CHANGELOG.md: %w", err)
	}

	changelogContents := string(changelogBytes)

	// Take everything under unreleased header (either "Unreleased" or "Development") and put it under the
	// new version header.

	now := time.Now()
	currentDate := fmt.Sprintf(
		"%s %s",
		humanize.Ordinal(now.Day()),
		now.Format("January 2006"),
	)

	replacement := fmt.Sprintf("## $1\n\n–\n\n## %s - %s", newVersion, currentDate)
	changelogContents = unreleasedHeaderRe.ReplaceAllString(changelogContents, replacement)

	_ = file.Close()

	file, err = os.OpenFile(c.filePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("error opening CHANGELOG.md for writing: %w", err)
	}
	defer file.Close()

	if _, err := file.WriteString(changelogContents); err != nil {
		return fmt.Errorf("error writing to CHANGELOG.md: %w", err)
	}

	return nil
}
