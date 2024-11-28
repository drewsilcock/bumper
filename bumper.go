package main

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/manifoldco/promptui"
	"github.com/rs/zerolog/log"
)

type BumpType int

const (
	BumpTypeMajor BumpType = iota
	BumpTypeMinor
	BumpTypePatch
)

type Bumper struct {
	conf *Config
}

func (b *Bumper) Bump() error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting current working directory: %w", err)
	}

	git := GitWrapper{}

	latestTag, err := git.GetLatestTag()
	if err != nil {
		return fmt.Errorf("error getting latest tag: %w", err)
	}

	packager := packagerForProject(cwd)
	if packager == nil {
		log.Warn().Msg("No supported package file found - package version will not be bumped")
	} else {
		if packager.Version() != "" && packager.Version() != latestTag {
			return fmt.Errorf("latest tag %s does not match package version %s", latestTag, packager.Version())
		}
	}

	changelogUpdater := NewChangelogUpdater(cwd)

	readmeParser := NewReadmeParser(cwd)
	projectName, err := readmeParser.GetProjectName()
	if err != nil {
		return fmt.Errorf("error getting project name: %w", err)
	}

	releaseCreator, err := getReleaseCreator(projectName, b.conf)
	if err != nil {
		return fmt.Errorf("error getting release creator: %w", err)
	}

	packagerName := "none"
	if packager != nil {
		packagerName = packager.Name()
	}
	log.Debug().Msgf(
		"Project info: name=%s, current version=%s, server=%s, packager=%s",
		projectName,
		latestTag,
		releaseCreator.Name(),
		packagerName,
	)

	currentBranch, err := git.GetCurrentBranch()
	if err != nil {
		return fmt.Errorf("error getting current branch: %w", err)
	}

	if currentBranch != "dev" {
		return fmt.Errorf("expected current branch to be 'dev', got %s", currentBranch)
	}

	if hasChanges, err := git.HasUncommittedChanges(); err != nil {
		return fmt.Errorf("error checking for uncommitted changes: %w", err)
	} else if hasChanges {
		return errors.New("uncommitted changes found - commit / stash changes before bumping version")
	}

	// Use the tag version as the last version as some packages don't contain versions.
	lastVersion, err := semver.NewVersion(latestTag)
	if err != nil {
		return fmt.Errorf("error parsing last version: %w", err)
	}

	majorBumpVersion := lastVersion.IncMajor()
	minorBumpVersion := lastVersion.IncMinor()
	patchBumpVersion := lastVersion.IncPatch()

	if b.conf.BumpType == nil {
		// Prompt for whether to do major, minor or patch bump.
		prompt := promptui.Select{
			Label: fmt.Sprintf("Select a version to bump to (current: v%s)", lastVersion),
			Items: []string{
				fmt.Sprintf("Major (v%s)", majorBumpVersion.String()),
				fmt.Sprintf("Minor (v%s)", minorBumpVersion.String()),
				fmt.Sprintf("Patch (v%s)", patchBumpVersion.String()),
			},
		}

		resultIndex, _, err := prompt.Run()
		if errors.Is(err, promptui.ErrInterrupt) {
			return fmt.Errorf("bump aborted")
		} else if err != nil {
			return fmt.Errorf("error selecting version bump: %w", err)
		}

		b.conf.BumpType = Ptr(BumpType(resultIndex))
	}

	var bumpVersion semver.Version
	var bumpText string
	switch *b.conf.BumpType {
	case BumpTypeMajor:
		bumpVersion = majorBumpVersion
		bumpText = "major"
	case BumpTypeMinor:
		bumpVersion = minorBumpVersion
		bumpText = "minor"
	case BumpTypePatch:
		bumpVersion = patchBumpVersion
		bumpText = "patch"
	default:
		return errors.New("invalid version bump selection")
	}

	newVersion := fmt.Sprintf("v%s", bumpVersion.String())
	log.Info().Msgf("Bumping %s version from %s to %s", bumpText, latestTag, newVersion)

	releaseBranchName := fmt.Sprintf("release/%s", newVersion)
	log.Debug().Msgf("Creating branch %s", releaseBranchName)
	if err := git.CreateBranch(releaseBranchName); err != nil {
		return fmt.Errorf("error creating release branch: %w", err)
	}

	if packager != nil {
		log.Debug().Msgf("Bumping package version from %s to %s", latestTag, newVersion)
		if err := packager.BumpVersion(newVersion); err != nil {
			return fmt.Errorf("error bumping package version: %w", err)
		}

		if err := git.Add(packager.PackageFilePath()); err != nil {
			return fmt.Errorf("error adding package file: %w", err)
		}
	} else {
		log.Debug().Msg("No supported package file found - skipping package version bump")
	}

	log.Debug().Msgf("Shifting unreleased changelog notes to %s", newVersion)
	if err := changelogUpdater.Update(newVersion); err != nil {
		return fmt.Errorf("error updating changelog: %w", err)
	}

	if err := git.Add(path.Join(cwd, "CHANGELOG.md")); err != nil {
		return fmt.Errorf("error adding changelog: %w", err)
	}

	// Now that we've updated the changelog, we can pull out the section for the new version.
	releaseNotes, err := changelogUpdater.GetVersionNotes(newVersion)
	if err != nil {
		return fmt.Errorf("error getting version notes: %w", err)
	}

	if !b.conf.Force {
		git.RunDiff(true)

		confirmPrompt := promptui.Prompt{
			Label:     fmt.Sprintf("About to bump version to %s - continue?", newVersion),
			IsConfirm: true,
		}

		shouldContinue, err := confirmPrompt.Run()
		if errors.Is(err, promptui.ErrAbort) || errors.Is(err, promptui.ErrInterrupt) || strings.ToLower(shouldContinue) != "y" {
			log.Debug().Msg("Cancelling bump")

			if err := git.RevertChanges(); err != nil {
				return fmt.Errorf("error reverting staged changes: %w", err)
			}

			if err := git.CheckoutBranch("dev"); err != nil {
				return fmt.Errorf("error switching back to dev branch: %w", err)
			}

			if err := git.DeleteBranch(releaseBranchName); err != nil {
				return fmt.Errorf("error deleting release branch: %w", err)
			}

			return errors.New("version bump cancelled")
		} else if err != nil {
			return fmt.Errorf("error confirming version bump: %w", err)
		}
	}

	log.Debug().Msg("Committing changes")
	if err := git.Commit(fmt.Sprintf("Bump version to %s", newVersion)); err != nil {
		return fmt.Errorf("error committing version bump: %w", err)
	}

	if err := git.CheckoutBranch("main"); err != nil {
		return fmt.Errorf("error switching to main branch: %w", err)
	}

	log.Debug().Msgf("Merging release branch %s into main", releaseBranchName)
	if err := git.MergeBranch(releaseBranchName); err != nil {
		return fmt.Errorf("error merging release branch: %w", err)
	}

	log.Debug().Msgf("Deleting release branch %s", releaseBranchName)
	if err := git.DeleteBranch(releaseBranchName); err != nil {
		return fmt.Errorf("error deleting release branch: %w", err)
	}

	log.Debug().Msgf("Creating tag %s", newVersion)
	if err := git.Tag(newVersion); err != nil {
		return fmt.Errorf("error tagging: %w", err)
	}

	log.Debug().Msg("Pushing commits")
	if err := git.Push(); err != nil {
		return fmt.Errorf("error pushing: %w", err)
	}

	log.Debug().Msg("Pushing tags")
	if err := git.PushTags(); err != nil {
		return fmt.Errorf("error pushing tags: %w", err)
	}

	if err := git.CheckoutBranch("dev"); err != nil {
		return fmt.Errorf("error switching to dev branch: %w", err)
	}

	log.Debug().Msg("Merging main into dev")
	if err := git.MergeBranch("main"); err != nil {
		return fmt.Errorf("error merging main branch: %w", err)
	}

	log.Debug().Msg("Pushing dev commits")
	if err := git.Push(); err != nil {
		return fmt.Errorf("error pushing: %w", err)
	}

	log.Debug().Msgf("Creating release in %s", releaseCreator.Name())
	releaseURL, err := releaseCreator.CreateRelease(newVersion, releaseNotes)
	if err != nil {
		return fmt.Errorf("error creating release: %w", err)
	}

	log.Info().Msgf("Created %s release: %s", releaseCreator.Name(), releaseURL.String())

	log.Info().Msgf("Successfully bumped version from %s to %s", latestTag, newVersion)
	return nil
}
