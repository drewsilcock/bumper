package main

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"regexp"

	"github.com/BurntSushi/toml"
	"github.com/Masterminds/semver"
	"github.com/rs/zerolog/log"
)

var tomlVersionRe = regexp.MustCompile(`(?m)^\s*version\s*=\s*"[^"]+"\s*$`)

type PyprojectPackager struct {
	packageFilePath string
	version         semver.Version
}

func (p *PyprojectPackager) Parse(projectPath string) error {
	packageFilePath := path.Join(projectPath, "pyproject.toml")

	if !fileExists(packageFilePath) {
		return ErrPackageNotFound
	}

	packageFile, err := os.Open(packageFilePath)
	if err != nil {
		return fmt.Errorf("error opening pyproject.toml: %w", err)
	}
	defer packageFile.Close()

	packageBytes, err := io.ReadAll(packageFile)
	if err != nil {
		return fmt.Errorf("error reading pyproject.toml: %w", err)
	}

	var packageContents map[string]interface{}
	if err := toml.Unmarshal(packageBytes, &packageContents); err != nil {
		return fmt.Errorf("error parsing pyproject.toml: %w", err)
	}

	// Try PEP-621 first as it's the modern standard, then fall back to Poetry's format.
	packageVersionRaw, err := tryParsePEP621(packageContents)
	if err != nil {
		packageVersionRaw, err = tryParsePoetry(packageContents)
		if err != nil {
			return errors.New("unable to find version in pyproject.toml")
		}
	}

	version, err := semver.NewVersion(packageVersionRaw)
	if err != nil {
		return errors.New("invalid semver version")
	}

	p.packageFilePath = packageFilePath
	p.version = *version

	return nil
}

func tryParsePoetry(packageContents map[string]interface{}) (string, error) {
	toolSection, ok := packageContents["tool"].(map[string]interface{})
	if !ok {
		return "", errors.New("tool section not found in pyproject.toml")
	}

	poetrySection, ok := toolSection["poetry"].(map[string]interface{})
	if !ok {
		return "", errors.New("poetry section not found in pyproject.toml")
	}

	packageVersionRaw, ok := poetrySection["version"].(string)
	if !ok {
		return "", errors.New("version not found in pyproject.toml")
	}

	return packageVersionRaw, nil
}

func tryParsePEP621(packageContents map[string]interface{}) (string, error) {
	projectSection, ok := packageContents["project"].(map[string]interface{})
	if !ok {
		return "", errors.New("project section not found in pyproject.toml")
	}

	packageVersionRaw, ok := projectSection["version"].(string)
	if !ok {
		return "", errors.New("version not found in pyproject.toml")
	}

	return packageVersionRaw, nil
}

func (p *PyprojectPackager) Name() string {
	return "poetry"
}

func (p *PyprojectPackager) Version() string {
	return fmt.Sprintf("v%s", p.version.String())
}

func (p *PyprojectPackager) PackageFilePath() string {
	return p.packageFilePath
}

func (p *PyprojectPackager) BumpVersion(newVersion string) error {
	packageFile, err := os.OpenFile(p.packageFilePath, os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("error opening pyproject.toml: %w", err)
	}
	defer func(packageFile *os.File) {
		if err := packageFile.Close(); err != nil && !errors.Is(err, fs.ErrClosed) {
			log.Error().Err(err).Msg("error closing pyproject.toml")
		}
	}(packageFile)

	packageBytes, err := io.ReadAll(packageFile)
	if err != nil {
		return fmt.Errorf("error reading pyproject.toml: %w", err)
	}

	if newVersion[0] == 'v' {
		newVersion = newVersion[1:]
	}
	replacementBytes := []byte(fmt.Sprintf("version = \"%s\"\n", newVersion))

	packageBytes = tomlVersionRe.ReplaceAll(packageBytes, replacementBytes)

	if err := packageFile.Close(); err != nil {
		return fmt.Errorf("error closing pyproject.toml: %w", err)
	}

	packageFile, err = os.OpenFile(p.packageFilePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("error opening pyproject.toml: %w", err)
	}
	defer func(packageFile *os.File) {
		if err := packageFile.Close(); err != nil {
			log.Error().Err(err).Msg("error closing pyproject.toml")
		}
	}(packageFile)

	if _, err := packageFile.Write(packageBytes); err != nil {
		return fmt.Errorf("error writing to pyproject.toml: %w", err)
	}

	return nil
}
