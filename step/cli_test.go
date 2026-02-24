package step

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildClaudeArgs_BasicPrompt(t *testing.T) {
	input := Input{
		Prompt:        "Say hello",
		APIKey:        "sk-test",
		ClaudeVersion: "2.1.45",
	}

	args, err := buildClaudeArgs(input, "")
	require.NoError(t, err)

	require.GreaterOrEqual(t, len(args), 2)
	assert.Equal(t, "--print", args[0])
	assert.Equal(t, "Say hello", args[1])
}

func TestBuildClaudeArgs_AllowedTools(t *testing.T) {
	input := Input{
		Prompt:       "hello",
		APIKey:       "sk-test",
		AllowedTools: "Read,Write,Edit",
	}

	args, err := buildClaudeArgs(input, "")
	require.NoError(t, err)

	assertContainsSequence(t, args, "--allowed-tools", "Read,Write,Edit")
}

func TestBuildClaudeArgs_NoAllowedTools(t *testing.T) {
	input := Input{
		Prompt: "hello",
		APIKey: "sk-test",
	}

	args, err := buildClaudeArgs(input, "")
	require.NoError(t, err)

	assert.NotContains(t, args, "--allowed-tools")
}

func TestBuildClaudeArgs_Agents(t *testing.T) {
	input := Input{
		Prompt: "hello",
		APIKey: "sk-test",
		Agents: "computer-use",
	}

	args, err := buildClaudeArgs(input, "")
	require.NoError(t, err)

	assertContainsSequence(t, args, "--agents", "computer-use")
}

func TestBuildClaudeArgs_AgentsNone(t *testing.T) {
	input := Input{
		Prompt: "hello",
		APIKey: "sk-test",
		Agents: "none",
	}

	args, err := buildClaudeArgs(input, "")
	require.NoError(t, err)

	assert.NotContains(t, args, "--agents")
}

func TestBuildClaudeArgs_MCPConfig(t *testing.T) {
	input := Input{
		Prompt: "hello",
		APIKey: "sk-test",
	}

	args, err := buildClaudeArgs(input, "/tmp/mcp-config.json")
	require.NoError(t, err)

	assertContainsSequence(t, args, "--mcp-config", "/tmp/mcp-config.json")
}

func TestBuildClaudeArgs_NoMCPConfig(t *testing.T) {
	input := Input{
		Prompt: "hello",
		APIKey: "sk-test",
	}

	args, err := buildClaudeArgs(input, "")
	require.NoError(t, err)

	assert.NotContains(t, args, "--mcp-config")
}

func TestBuildClaudeArgs_AdditionalFlags(t *testing.T) {
	input := Input{
		Prompt:             "hello",
		APIKey:             "sk-test",
		AdditionalCLIFlags: "--max-turns 5 --verbose",
	}

	args, err := buildClaudeArgs(input, "")
	require.NoError(t, err)

	assertContainsSequence(t, args, "--max-turns", "5")
	assert.Contains(t, args, "--verbose")
}

func TestBuildClaudeArgs_AdditionalFlagsQuoted(t *testing.T) {
	input := Input{
		Prompt:             "hello",
		APIKey:             "sk-test",
		AdditionalCLIFlags: `--system-prompt "You are helpful"`,
	}

	args, err := buildClaudeArgs(input, "")
	require.NoError(t, err)

	assertContainsSequence(t, args, "--system-prompt", "You are helpful")
}

func TestBuildClaudeArgs_AdditionalFlagsUnclosedQuote(t *testing.T) {
	input := Input{
		Prompt:             "hello",
		APIKey:             "sk-test",
		AdditionalCLIFlags: `--flag "unclosed`,
	}

	_, err := buildClaudeArgs(input, "")
	require.Error(t, err)
}

// assertContainsSequence checks that a, b appear as consecutive elements in haystack.
func assertContainsSequence(t *testing.T, haystack []string, a, b string) {
	t.Helper()
	for i := 0; i+1 < len(haystack); i++ {
		if haystack[i] == a && haystack[i+1] == b {
			return
		}
	}
	assert.Failf(t, "sequence not found", "expected [%q, %q] as consecutive elements in %v", a, b, haystack)
}
