package main

import (
	"os"

	"github.com/bitrise-io/go-steputils/v2/stepconf"
	"github.com/bitrise-io/go-utils/v2/command"
	"github.com/bitrise-io/go-utils/v2/env"
	"github.com/bitrise-io/go-utils/v2/errorutil"
	"github.com/bitrise-io/go-utils/v2/exitcode"
	"github.com/bitrise-io/go-utils/v2/log"

	"github.com/bitrise-steplib/bitrise-step-claude-code/step"
)

func main() {
	os.Exit(int(run()))
}

func run() exitcode.ExitCode {
	logger := log.NewLogger()
	envRepo := env.NewRepository()
	inputParser := stepconf.NewInputParser(envRepo)
	commandFactory := command.NewFactory(envRepo)

	s := step.New(logger, inputParser, commandFactory, envRepo)
	if err := s.Run(); err != nil {
		logger.Errorf("%s", errorutil.FormattedError(err))
		return exitcode.Failure
	}
	return exitcode.Success
}
