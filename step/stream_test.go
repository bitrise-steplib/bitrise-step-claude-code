package step

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const sampleNDJSON = `{"type":"system","subtype":"init","tools":["Bash","Read"]}
{"type":"assistant","message":{"content":[{"type":"text","text":"Let me run that for you."}]}}
{"type":"assistant","message":{"content":[{"type":"tool_use","name":"Bash","input":{"command":"echo hello"}}]}}
{"type":"user","message":{"content":[{"type":"tool_result","content":"hello\n","is_error":false}]}}
{"type":"result","result":"Done","is_error":false,"num_turns":2,"duration_ms":1234,"total_cost_usd":0.0012}`

func TestParseStream_Success(t *testing.T) {
	logger := log.NewLogger()
	output, err := parseStream(strings.NewReader(sampleNDJSON), logger)
	require.NoError(t, err)
	assert.Equal(t, "Done", output)
}

func TestParseStream_IsError(t *testing.T) {
	ndjson := `{"type":"result","result":"","is_error":true,"num_turns":1,"duration_ms":500,"total_cost_usd":0}`
	logger := log.NewLogger()
	_, err := parseStream(strings.NewReader(ndjson), logger)
	require.Error(t, err)
}

func TestToolResultFirstLine(t *testing.T) {
	content := `"Exit code 1\ngh: pull request not found\n"`
	assert.Equal(t, "Exit code 1", toolResultFirstLine(json.RawMessage(content)))
}

func TestToolResultText_MultiLine(t *testing.T) {
	content := `"Exit code 1\ngh: pull request not found\n"`
	assert.Equal(t, "Exit code 1\ngh: pull request not found", toolResultText(json.RawMessage(content)))
}

func TestParseStream_NoResult(t *testing.T) {
	ndjson := `{"type":"assistant","message":{"content":[{"type":"text","text":"hi"}]}}`
	logger := log.NewLogger()
	_, err := parseStream(strings.NewReader(ndjson), logger)
	require.Error(t, err)
}

func TestParseStream_IsError_IncludesResultText(t *testing.T) {
	ndjson := `{"type":"result","result":"Permission denied","is_error":true,"num_turns":1,"duration_ms":500,"total_cost_usd":0}`
	logger := log.NewLogger()
	_, err := parseStream(strings.NewReader(ndjson), logger)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Permission denied")
}

func TestToolResultText_ArrayContent(t *testing.T) {
	content := `[{"type":"text","text":"line one"},{"type":"text","text":"line two"}]`
	assert.Equal(t, "line one\nline two", toolResultText(json.RawMessage(content)))
}

func TestToolResultFirstLine_ArrayContent(t *testing.T) {
	content := `[{"type":"text","text":"first line"},{"type":"text","text":"second line"}]`
	assert.Equal(t, "first line", toolResultFirstLine(json.RawMessage(content)))
}

func TestFormatToolInput(t *testing.T) {
	tests := []struct {
		name     string
		toolName string
		input    string
		want     string
	}{
		{"bash command", "Bash", `{"command":"echo hello"}`, "echo hello"},
		{"read file", "Read", `{"file_path":"/tmp/foo.go"}`, "/tmp/foo.go"},
		{"write file", "Write", `{"file_path":"/tmp/bar.go","content":"x"}`, "/tmp/bar.go"},
		{"edit file", "Edit", `{"file_path":"/tmp/baz.go"}`, "/tmp/baz.go"},
		{"glob pattern", "Glob", `{"pattern":"**/*.go"}`, "**/*.go"},
		{"grep pattern", "Grep", `{"pattern":"func main"}`, "func main"},
		{"unknown tool", "Task", `{"prompt":"do something"}`, ""},
		{"empty input", "Bash", `{}`, ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := formatToolInput(tc.toolName, json.RawMessage(tc.input))
			assert.Equal(t, tc.want, got)
		})
	}
}
