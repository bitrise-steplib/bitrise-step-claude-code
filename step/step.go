package step

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/bitrise-io/go-steputils/v2/stepconf"
	"github.com/bitrise-io/go-utils/v2/command"
	"github.com/bitrise-io/go-utils/v2/env"
	"github.com/bitrise-io/go-utils/v2/log"
)

type Input struct {
	Prompt             string `env:"prompt,required"`
	APIKey             string `env:"api_key,required"`
	AllowedTools       string `env:"allowed_tools"`
	Agents             string `env:"agents"`
	BitriseToken       string `env:"bitrise_token"`
	AdditionalCLIFlags string `env:"additional_cli_flags"`
	ClaudeVersion      string `env:"claude_version,required"`
}

type Step struct {
	logger         log.Logger
	inputParser    stepconf.InputParser
	commandFactory command.Factory
	envRepo        env.Repository
	exporter       outputExporter
}

func New(logger log.Logger, inputParser stepconf.InputParser, commandFactory command.Factory, envRepo env.Repository) *Step {
	return &Step{
		logger:         logger,
		inputParser:    inputParser,
		commandFactory: commandFactory,
		envRepo:        envRepo,
		exporter:       newExporter(commandFactory),
	}
}

func (s *Step) Run() error {
	var input Input
	if err := s.inputParser.Parse(&input); err != nil {
		return fmt.Errorf("parse inputs: %w", err)
	}

	stepconf.Print(input)

	if err := s.installClaude(input.ClaudeVersion); err != nil {
		return fmt.Errorf("install claude: %w", err)
	}

	// Add ~/.local/bin to PATH so the installed claude binary is found.
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("get home dir: %w", err)
	}
	localBin := filepath.Join(homeDir, ".local", "bin")
	currentPath := os.Getenv("PATH")
	if err := os.Setenv("PATH", localBin+string(os.PathListSeparator)+currentPath); err != nil {
		return fmt.Errorf("update PATH: %w", err)
	}

	var mcpConfigPath string
	if input.BitriseToken != "" {
		mcpConfigPath, err = writeMCPConfig(input.BitriseToken)
		if err != nil {
			return fmt.Errorf("write MCP config: %w", err)
		}
		defer os.Remove(mcpConfigPath)
	}

	args, err := buildClaudeArgs(input, mcpConfigPath)
	if err != nil {
		return fmt.Errorf("build claude args: %w", err)
	}

	output, err := s.runClaude(args, input.APIKey)
	if err != nil {
		return fmt.Errorf("run claude: %w", err)
	}

	if err := s.exportOutputs(output); err != nil {
		return fmt.Errorf("export outputs: %w", err)
	}

	return nil
}
