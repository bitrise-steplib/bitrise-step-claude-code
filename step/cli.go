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

// buildClaudeArgs leaves out the prompt, which goes to stdin instead. As an argument it would be
// parsed as an option when starting with "-", hit the 128 KiB per-argument limit on Linux and be
// echoed in command error messages.
func buildClaudeArgs(input Input, mcpConfigPath string) ([]string, error) {
	var args []string
	if input.LogFormat == "raw" {
		// --print: non-interactive mode that buffers and prints the final response.
		// No JSON parsing needed — safe fallback if the stream-json format ever changes.
		args = []string{"--print"}
	} else {
		// --output-format stream-json enables non-interactive mode with real-time NDJSON events,
		// giving visibility into tool calls and assistant text as they happen.
		// --verbose is required alongside --output-format stream-json (CLI enforces this).
		args = []string{"--output-format", "stream-json", "--verbose"}
	}

	if input.Model != "" {
		args = append(args, "--model", input.Model)
	}

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

func (s *Step) runClaude(args []string, input Input) (string, error) {
	s.logger.Infof("Running Claude...")
	s.logger.Println()

	if input.LogFormat == "raw" {
		return s.runClaudeRaw(args, input.Prompt, input.APIKey)
	}
	return s.runClaudeStreaming(args, input.Prompt, input.APIKey)
}

func (s *Step) runClaudeStreaming(args []string, prompt, apiKey string) (string, error) {
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
		Stdin:  strings.NewReader(prompt),
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

func (s *Step) runClaudeRaw(args []string, prompt, apiKey string) (string, error) {
	cmd := s.commandFactory.Create("claude", args, &command.Opts{
		Stdin: strings.NewReader(prompt),
		Env:   []string{"ANTHROPIC_API_KEY=" + apiKey},
	})

	output, err := cmd.RunAndReturnTrimmedCombinedOutput()
	if err != nil {
		return "", fmt.Errorf("claude exited with error: %w", err)
	}
	return output, nil
}
