package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"regexp"

	"github.com/Masterminds/semver"
	"github.com/rs/zerolog/log"
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
	defer func(packageFile *os.File) {
		if err := packageFile.Close(); err != nil {
			log.Error().Err(err).Msg("error closing package.json")
		}
	}(packageFile)

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
	return fmt.Sprintf("v%s", p.version.String())
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
	defer func(packageFile *os.File) {
		if err := packageFile.Close(); err != nil && !errors.Is(err, fs.ErrClosed) {
			log.Error().Err(err).Msg("error closing package.json")
		}
	}(packageFile)

	packageBytes, err := io.ReadAll(packageFile)
	if err != nil {
		return fmt.Errorf("error reading package.json: %w", err)
	}

	packageContents := string(packageBytes)

	// We don't want the 'v' in the package.json version.
	if newVersion[0] == 'v' {
		newVersion = newVersion[1:]
	}
	replacementStr := fmt.Sprintf("$1\"version\": \"%s\"\n", newVersion)

	packageContents = packageVersionRe.ReplaceAllString(packageContents, replacementStr)

	if err := packageFile.Close(); err != nil {
		return fmt.Errorf("error closing package.json: %w", err)
	}

	packageFile, err = os.OpenFile(p.packageFilePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("error opening package.json for writing: %w", err)
	}
	defer func(packageFile *os.File) {
		err := packageFile.Close()
		if err != nil {
			log.Error().Err(err).Msg("error closing package.json")
		}
	}(packageFile)

	if _, err = packageFile.WriteString(packageContents); err != nil {
		return fmt.Errorf("error writing package.json: %w", err)
	}

	return nil
}
