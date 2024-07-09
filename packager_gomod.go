package main

import "path"

type GoModPackager struct {
	packageFilePath string
}

// Some packages do put the version in the package URL, but we're not going to worry about parsing
// and bump that at the moment, so these are all a no-op.

func (p *GoModPackager) Parse(projectPath string) error {
	p.packageFilePath = path.Join(projectPath, "go.mod")

	if !fileExists(p.packageFilePath) {
		return ErrPackageNotFound
	}

	return nil
}

func (p *GoModPackager) Name() string {
	return "go mod"
}

func (p *GoModPackager) Version() string {
	return ""
}

func (p *GoModPackager) PackageFilePath() string {
	return p.packageFilePath
}

func (p *GoModPackager) BumpVersion(newVersion string) error {
	return nil
}
