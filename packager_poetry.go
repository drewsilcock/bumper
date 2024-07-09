package main

import (
	"errors"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/Masterminds/semver"
	"io"
	"os"
	"path"
	"regexp"
)

var tomlVersionRe = regexp.MustCompile(`^(\s+)version\s*=\s*"[^"]+"\s*$`)

type PoetryPackager struct {
	packageFilePath string
	version         semver.Version
}

func (p *PoetryPackager) Parse(projectPath string) error {
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

	toolSection, ok := packageContents["tool"].(map[string]interface{})
	if !ok {
		return errors.New("tool section not found in pyproject.toml")
	}

	poetrySection, ok := toolSection["poetry"].(map[string]interface{})
	if !ok {
		return errors.New("poetry section not found in pyproject.toml")
	}

	packageVersionRaw, ok := poetrySection["version"].(string)
	if !ok {
		return errors.New("version not found in pyproject.toml")
	}

	version, err := semver.NewVersion(packageVersionRaw)
	if err != nil {
		return errors.New("invalid semver version")
	}

	p.packageFilePath = packageFilePath
	p.version = *version

	return nil
}

func (p *PoetryPackager) Name() string {
	return "poetry"
}

func (p *PoetryPackager) Version() string {
	return p.version.String()
}

func (p *PoetryPackager) PackageFilePath() string {
	return p.packageFilePath
}

func (p *PoetryPackager) BumpVersion(newVersion string) error {
	packageFile, err := os.OpenFile(p.packageFilePath, os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("error opening pyproject.toml: %w", err)
	}
	defer packageFile.Close()

	packageBytes, err := io.ReadAll(packageFile)
	if err != nil {
		return fmt.Errorf("error reading pyproject.toml: %w", err)
	}

	replacementBytes := []byte(fmt.Sprintf(`${1}version = "%s"`, newVersion))

	packageBytes = tomlVersionRe.ReplaceAll(packageBytes, replacementBytes)

	_ = packageFile.Close()

	packageFile, err = os.OpenFile(p.packageFilePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("error opening pyproject.toml: %w", err)
	}
	defer packageFile.Close()

	if _, err := packageFile.Write(packageBytes); err != nil {
		return fmt.Errorf("error writing to pyproject.toml: %w", err)
	}

	return nil
}
