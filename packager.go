package main

import (
	"errors"
	"os"
)

var ErrPackageNotFound = errors.New("package file(s) not found")

type Packager interface {
	Parse(projectPath string) error
	Name() string
	Version() string
	PackageFilePath() string
	BumpVersion(newVersion string) error
}

func packagerForProject(projectPath string) Packager {
	packagers := []Packager{
		&GoModPackager{},
		&PoetryPackager{},
		&NPMPackager{},
	}

	for _, packager := range packagers {
		if err := packager.Parse(projectPath); err == nil {
			return packager
		}
	}

	return nil
}

func fileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return !errors.Is(err, os.ErrNotExist)
}
