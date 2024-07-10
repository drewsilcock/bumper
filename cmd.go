package main

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "bumper",
		Short: "Drew's version bumper",
		Long:  "Tool for bumping versions using git flow and GitLab / GitHub releases.",
		Run:   run,
	}

	args Args
)

type Args struct {
	BumpType string
	Force    bool
}

func ExecuteCmd() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(Setup)

	rootCmd.PersistentFlags().StringVarP(
		&args.BumpType,
		"type",
		"t",
		"",
		"type of version bump (major, minor, patch) [optional]",
	)

	rootCmd.PersistentFlags().BoolVarP(
		&args.Force,
		"force",
		"f",
		false,
		"run without confirmation [optional]",
	)
}

func run(_ *cobra.Command, _ []string) {
	conf := NewConfig(args)

	bumper := Bumper{conf: conf}
	if err := bumper.Bump(); err != nil {
		log.Fatal().Msgf("Failed to bump version: %v", err)
	}
}
