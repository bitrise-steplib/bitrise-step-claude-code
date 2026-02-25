package step

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/bitrise-io/go-utils/v2/command"
	shellquote "github.com/kballard/go-shellquote"
)

func buildClaudeArgs(input Input, mcpConfigPath string) ([]string, error) {
	// --output-format stream-json enables non-interactive mode with real-time NDJSON events,
	// giving visibility into tool calls and assistant text as they happen.
	// --verbose is required alongside --output-format stream-json (CLI enforces this).
	// The prompt is the positional [prompt] argument.
	args := []string{"--output-format", "stream-json", "--verbose", input.Prompt}

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
	s.logger.Println()

	pipeReader, pipeWriter := io.Pipe()

	type streamResult struct {
		output string
		err    error
	}
	ch := make(chan streamResult, 1)

	go func() {
		output, err := parseStream(pipeReader, s.logger)
		ch <- streamResult{output, err}
	}()

	// Tee stderr to both os.Stderr (so it appears in the build log in real-time) and a
	// buffer. The buffer lets us attach the CLI's error message to the returned Go error —
	// otherwise the cause ends up floating somewhere earlier in the log, disconnected from
	// the error that actually fails the step. This matters especially for CLI startup errors
	// (bad flags, version mismatches, etc.) that exit before emitting any NDJSON events,
	// which would otherwise surface only as the generic "no result received from Claude".
	var stderrBuf bytes.Buffer
	cmd := s.commandFactory.Create("claude", args, &command.Opts{
		Stdout: pipeWriter,
		Stderr: io.MultiWriter(os.Stderr, &stderrBuf),
		Env:    []string{"ANTHROPIC_API_KEY=" + apiKey},
	})

	if err := cmd.Start(); err != nil {
		pipeWriter.CloseWithError(err)
		<-ch
		return "", fmt.Errorf("start claude: %w", err)
	}

	cmdErr := cmd.Wait()
	pipeWriter.Close()

	result := <-ch

	if result.err != nil {
		if stderrBuf.Len() > 0 {
			return "", fmt.Errorf("%w\n%s", result.err, strings.TrimSpace(stderrBuf.String()))
		}
		return "", result.err
	}
	if cmdErr != nil {
		if stderrBuf.Len() > 0 {
			return "", fmt.Errorf("claude exited with error: %w\n%s", cmdErr, strings.TrimSpace(stderrBuf.String()))
		}
		return "", fmt.Errorf("claude exited with error: %w", cmdErr)
	}

	return result.output, nil
}
