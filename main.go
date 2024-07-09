package main

import (
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339})

	bumper := Bumper{ShouldConfirm: true}
	if err := bumper.Bump(); err != nil {
		log.Fatal().Msg(fmt.Sprintf("Failed to bump version: %v", err))
	}
}
