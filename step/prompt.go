package step

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var envVarRef = regexp.MustCompile(`\$(?:\{[A-Za-z_][A-Za-z0-9_]*\}|[A-Za-z_][A-Za-z0-9_]*)`)

// resolvePrompt returns the prompt text and a loggable description of where it came from.
// Relative prompt_file paths resolve against the working directory, which is the source dir on Bitrise.
func resolvePrompt(input Input, lookupEnv func(string) (string, bool)) (prompt string, source string, err error) {
	switch {
	case input.Prompt != "" && input.PromptFile != "":
		return "", "", errors.New("prompt and prompt_file are mutually exclusive, set only one of them")
	case input.Prompt != "":
		return input.Prompt, "inline prompt input", nil
	case input.PromptFile == "":
		return "", "", errors.New("either prompt or prompt_file must be set")
	}

	path, err := filepath.Abs(input.PromptFile)
	if err != nil {
		return "", "", fmt.Errorf("resolve prompt_file path %s: %w", input.PromptFile, err)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return "", "", fmt.Errorf("read prompt_file: %w", err)
	}
	prompt = expandSetEnvVars(string(content), lookupEnv)
	if strings.TrimSpace(prompt) == "" {
		return "", "", fmt.Errorf("prompt_file %s is empty", path)
	}

	return prompt, "file " + path, nil
}

// expandSetEnvVars replaces $NAME and ${NAME} references with the variable's value when it is set.
// References to unset variables and any other "$" text, such as $1 or ${NAME:-default}, are left as is.
func expandSetEnvVars(s string, lookupEnv func(string) (string, bool)) string {
	return envVarRef.ReplaceAllStringFunc(s, func(ref string) string {
		if value, ok := lookupEnv(strings.Trim(ref, "${}")); ok {
			return value
		}
		return ref
	})
}
