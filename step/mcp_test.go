package step

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMCPPresets_Exist(t *testing.T) {
	tests := []struct {
		name    string
		wantCmd string
	}{
		{"bitrise", ""},       // SSE type, no command
		{"xcodebuild", "npx"},
		{"mobile-mcp", "npx"},
}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			preset, ok := mcpPresets[tt.name]
			require.True(t, ok, "preset %q should exist", tt.name)
			assert.Equal(t, tt.wantCmd, preset.Command)
		})
	}
}

func TestMCPPresets_BitriseIsSSE(t *testing.T) {
	preset := mcpPresets["bitrise"]
	assert.Equal(t, "sse", preset.Type)
	assert.Equal(t, "https://mcp.bitrise.io", preset.URL)
}

func TestBuildMCPConfig_SingleServer(t *testing.T) {
	servers := map[string]mcpServerConfig{
		"xcodebuild": {
			Command: "npx",
			Args:    []string{"xcodebuildmcp@latest", "mcp"},
		},
	}

	cfg := buildMCPConfig(servers)
	assert.Len(t, cfg.McpServers, 1)
	assert.Equal(t, "npx", cfg.McpServers["xcodebuild"].Command)
}

func TestBuildMCPConfig_MergesMultiple(t *testing.T) {
	servers := map[string]mcpServerConfig{
		"bitrise": {
			Type: "sse",
			URL:  "https://mcp.bitrise.io",
			Headers: map[string]string{
				"Authorization": "Bearer tok",
			},
		},
		"xcodebuild": {
			Command: "npx",
			Args:    []string{"xcodebuildmcp@latest", "mcp"},
		},
	}

	cfg := buildMCPConfig(servers)
	assert.Len(t, cfg.McpServers, 2)
}

func TestBuildMCPConfig_Empty(t *testing.T) {
	cfg := buildMCPConfig(nil)
	assert.Empty(t, cfg.McpServers)
}

func TestWriteMCPConfigFile(t *testing.T) {
	servers := map[string]mcpServerConfig{
		"xcodebuild": {Command: "npx", Args: []string{"xcodebuildmcp@latest", "mcp"}},
	}

	path, err := writeMCPConfigFile(buildMCPConfig(servers))
	require.NoError(t, err)
	defer os.Remove(path)

	data, err := os.ReadFile(path)
	require.NoError(t, err)

	var parsed mcpConfig
	require.NoError(t, json.Unmarshal(data, &parsed))
	assert.Equal(t, "npx", parsed.McpServers["xcodebuild"].Command)
}

func TestResolveMCPServers_WithToken(t *testing.T) {
	agent := AgentDefinition{
		RequiredMCPs: map[Platform][]string{
			PlatformApple: {"xcodebuild"},
		},
	}

	servers := resolveMCPServers(agent, PlatformApple, "tok")

	assert.Contains(t, servers, "bitrise")
	assert.Contains(t, servers, "xcodebuild")
}

func TestResolveMCPServers_NoToken(t *testing.T) {
	agent := AgentDefinition{
		RequiredMCPs: map[Platform][]string{
			PlatformApple: {"xcodebuild"},
		},
	}

	servers := resolveMCPServers(agent, PlatformApple, "")
	assert.NotContains(t, servers, "bitrise")
	assert.Contains(t, servers, "xcodebuild")
}

func TestResolveMCPServers_PlatformFiltering(t *testing.T) {
	agent := AgentDefinition{
		RequiredMCPs: map[Platform][]string{
			PlatformApple:   {"xcodebuild"},
			PlatformAndroid: {"mobile-mcp"},
			PlatformOther:   {"mobile-mcp"},
		},
	}

	appleServers := resolveMCPServers(agent, PlatformApple, "")
	assert.Contains(t, appleServers, "xcodebuild")
	assert.NotContains(t, appleServers, "mobile-mcp")

	androidServers := resolveMCPServers(agent, PlatformAndroid, "")
	assert.Contains(t, androidServers, "mobile-mcp")
	assert.NotContains(t, androidServers, "xcodebuild")

	otherServers := resolveMCPServers(agent, PlatformOther, "")
	assert.Contains(t, otherServers, "mobile-mcp")
	assert.NotContains(t, otherServers, "xcodebuild")
}

func TestResolveMCPServers_NoAgentMCPs(t *testing.T) {
	agent := AgentDefinition{}

	servers := resolveMCPServers(agent, PlatformApple, "tok")
	assert.Contains(t, servers, "bitrise", "bitrise still added with token")
	assert.Len(t, servers, 1)
}

func TestResolveMCPServers_BitriseRequiredButNoToken(t *testing.T) {
	agent := AgentDefinition{
		RequiredMCPs: map[Platform][]string{
			PlatformApple: {"bitrise", "xcodebuild"},
		},
	}

	servers := resolveMCPServers(agent, PlatformApple, "")
	assert.NotContains(t, servers, "bitrise", "bitrise should not be added without token")
	assert.Contains(t, servers, "xcodebuild", "non-auth MCPs still added")
}

func TestResolveMCPServers_NoAgentNoToken(t *testing.T) {
	agent := AgentDefinition{}

	servers := resolveMCPServers(agent, PlatformApple, "")
	assert.Empty(t, servers)
}
