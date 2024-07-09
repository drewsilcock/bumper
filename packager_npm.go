package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"regexp"

	"github.com/Masterminds/semver"
)

var packageVersionRe = regexp.MustCompile(`(?m)^(\s+)"version":\s*"[^"]+"\s*$`)

type NPMPackager struct {
	packageFilePath string
	version         semver.Version
}

func (p *NPMPackager) Parse(projectPath string) error {
	packageFilePath := path.Join(projectPath, "package.json")

	if !fileExists(packageFilePath) {
		return ErrPackageNotFound
	}

	packageFile, err := os.Open(packageFilePath)
	if err != nil {
		return fmt.Errorf("error opening package.json: %w", err)
	}
	defer packageFile.Close()

	packageBytes, err := io.ReadAll(packageFile)
	if err != nil {
		return fmt.Errorf("error reading package.json: %w", err)
	}

	packageContents := make(map[string]interface{})
	err = json.Unmarshal(packageBytes, &packageContents)
	if err != nil {
		return fmt.Errorf("error parsing package.json: %w", err)
	}

	packageVersionRaw, ok := packageContents["version"].(string)
	if !ok {
		return errors.New("version not found in package.json")
	}

	packageVersion, err := semver.NewVersion(packageVersionRaw)
	if err != nil {
		return errors.New("invalid semver version")
	}

	p.packageFilePath = packageFilePath
	p.version = *packageVersion

	return nil
}

func (p *NPMPackager) Name() string {
	return "npm"
}

func (p *NPMPackager) Version() string {
	return p.version.String()
}

func (p *NPMPackager) PackageFilePath() string {
	return p.packageFilePath
}

func (p *NPMPackager) BumpVersion(newVersion string) error {
	// When we actually do the bump, just use regex replace to prevent messing up the formatting, order, etc.
	packageFile, err := os.Open(p.packageFilePath)
	if err != nil {
		return fmt.Errorf("error opening package.json: %w", err)
	}
	defer packageFile.Close()

	packageBytes, err := io.ReadAll(packageFile)
	if err != nil {
		return fmt.Errorf("error reading package.json: %w", err)
	}

	packageContents := string(packageBytes)

	replacementStr := fmt.Sprintf(`$1"version": "%s"`, newVersion)

	packageContents = packageVersionRe.ReplaceAllString(packageContents, replacementStr)

	_ = packageFile.Close()

	packageFile, err = os.OpenFile(p.packageFilePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("error opening package.json for writing: %w", err)
	}

	if _, err = packageFile.Write(packageBytes); err != nil {
		return fmt.Errorf("error writing package.json: %w", err)
	}

	return nil
}
