package step

import (
	"fmt"

	"github.com/bitrise-io/go-utils/v2/command"
	shellquote "github.com/kballard/go-shellquote"
)

func buildClaudeArgs(input Input, mcpConfigPath string) ([]string, error) {
	// --print enables non-interactive mode (print result and exit).
	// The prompt is the positional [prompt] argument, not a value of --print.
	args := []string{"--print", input.Prompt}

	if input.AllowedTools != "" {
		args = append(args, "--allowed-tools", input.AllowedTools)
	}

	if input.Agents != "" && input.Agents != "none" {
		args = append(args, "--agents", input.Agents)
	}

	if mcpConfigPath != "" {
		args = append(args, "--mcp-config", mcpConfigPath)
	}

	if input.AdditionalCLIFlags != "" {
		extra, err := shellquote.Split(input.AdditionalCLIFlags)
		if err != nil {
			return nil, fmt.Errorf("parse additional_cli_flags: %w", err)
		}
		args = append(args, extra...)
	}

	return args, nil
}

func (s *Step) runClaude(args []string, apiKey string) (string, error) {
	s.logger.Infof("Running Claude...")

	cmd := s.commandFactory.Create("claude", args, &command.Opts{
		Env: []string{"ANTHROPIC_API_KEY=" + apiKey},
	})

	output, err := cmd.RunAndReturnTrimmedCombinedOutput()
	if err != nil {
		return "", fmt.Errorf("claude exited with error: %w", err)
	}

	return output, nil
}
