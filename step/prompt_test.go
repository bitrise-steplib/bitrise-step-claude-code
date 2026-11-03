package step

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func lookupFrom(vars map[string]string) func(string) (string, bool) {
	return func(name string) (string, bool) {
		value, ok := vars[name]
		return value, ok
	}
}

func writePromptFile(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "prompt.md")
	require.NoError(t, os.WriteFile(path, []byte(content), 0o600))
	return path
}

func TestResolvePrompt_Inline(t *testing.T) {
	prompt, source, err := resolvePrompt(Input{Prompt: "Say $NAME"}, lookupFrom(map[string]string{"NAME": "hi"}))

	require.NoError(t, err)
	assert.Equal(t, "Say $NAME", prompt, "inline prompt is passed through unchanged")
	assert.Equal(t, "inline prompt input", source)
}

func TestResolvePrompt_File(t *testing.T) {
	path := writePromptFile(t, "Summarize changes on $BRANCH")

	prompt, source, err := resolvePrompt(Input{PromptFile: path}, lookupFrom(map[string]string{"BRANCH": "main"}))

	require.NoError(t, err)
	assert.Equal(t, "Summarize changes on main", prompt)
	assert.Equal(t, "file "+path, source)
}

func TestResolvePrompt_InvalidInputs(t *testing.T) {
	tests := []struct {
		name    string
		input   Input
		wantErr string
	}{
		{
			name:    "both set",
			input:   Input{Prompt: "hello", PromptFile: writePromptFile(t, "hello")},
			wantErr: "mutually exclusive",
		},
		{
			name:    "neither set",
			input:   Input{},
			wantErr: "either prompt or prompt_file must be set",
		},
		{
			name:    "missing file",
			input:   Input{PromptFile: filepath.Join(t.TempDir(), "missing.md")},
			wantErr: "read prompt_file",
		},
		{
			name:    "empty file",
			input:   Input{PromptFile: writePromptFile(t, "")},
			wantErr: "is empty",
		},
		{
			name:    "whitespace only file",
			input:   Input{PromptFile: writePromptFile(t, " \n\t\n")},
			wantErr: "is empty",
		},
		{
			name:    "file empty after expansion",
			input:   Input{PromptFile: writePromptFile(t, "$EMPTY\n")},
			wantErr: "is empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := resolvePrompt(tt.input, lookupFrom(map[string]string{"EMPTY": ""}))
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestExpandSetEnvVars(t *testing.T) {
	vars := map[string]string{"SET": "value", "EMPTY": ""}

	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "set var", in: "a $SET b", want: "a value b"},
		{name: "set var in braces", in: "a ${SET}b", want: "a valueb"},
		{name: "set but empty var", in: "a $EMPTY b", want: "a  b"},
		{name: "unset var", in: "a $UNSET b", want: "a $UNSET b"},
		{name: "unset var in braces", in: "a ${UNSET} b", want: "a ${UNSET} b"},
		{name: "shell default syntax", in: "${SET:-x} ${UNSET:-x}", want: "${SET:-x} ${UNSET:-x}"},
		{name: "positional and special params", in: "awk '{print $1}' $$ $@", want: "awk '{print $1}' $$ $@"},
		{name: "lone and malformed dollars", in: "costs $5, ${} and ${ end $", want: "costs $5, ${} and ${ end $"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, expandSetEnvVars(tt.in, lookupFrom(vars)))
		})
	}
}
