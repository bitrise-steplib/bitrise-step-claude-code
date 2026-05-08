package step

import (
	"encoding/json"
	"fmt"
	"os"
)

// mcpServerConfig represents a single MCP server entry.
// Supports both SSE-based servers (Type+URL+Headers) and stdio servers (Command+Args).
type mcpServerConfig struct {
	Type    string            `json:"type,omitempty"`
	URL     string            `json:"url,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
	Command string            `json:"command,omitempty"`
	Args    []string          `json:"args,omitempty"`
}

type mcpConfig struct {
	McpServers map[string]mcpServerConfig `json:"mcpServers"`
}

// mcpPresets contains the built-in MCP server configurations.
var mcpPresets = map[string]mcpServerConfig{
	"bitrise": {
		Type: "sse",
		URL:  "https://mcp.bitrise.io",
		// Headers are set dynamically with the auth token.
	},
	"xcodebuild": {
		Command: "npx",
		Args:    []string{"xcodebuildmcp@latest", "mcp"},
	},
	"mobile-mcp": {
		Command: "npx",
		Args:    []string{"-y", "@anthropic/mobile-mcp"},
	},
}

// buildMCPConfig creates an mcpConfig from a map of server configs.
func buildMCPConfig(servers map[string]mcpServerConfig) mcpConfig {
	if servers == nil {
		return mcpConfig{McpServers: map[string]mcpServerConfig{}}
	}
	return mcpConfig{McpServers: servers}
}

// writeMCPConfigFile writes the config to a temp file and returns the path.
func writeMCPConfigFile(cfg mcpConfig) (string, error) {
	data, err := json.Marshal(cfg)
	if err != nil {
		return "", fmt.Errorf("marshal MCP config: %w", err)
	}

	f, err := os.CreateTemp("", "claude-mcp-config-*.json")
	if err != nil {
		return "", fmt.Errorf("create temp file: %w", err)
	}
	defer f.Close()

	if _, err := f.Write(data); err != nil {
		os.Remove(f.Name())
		return "", fmt.Errorf("write MCP config: %w", err)
	}

	return f.Name(), nil
}

// resolveMCPServers determines which MCP servers to enable based on
// agent requirements, platform, and bitrise token.
func resolveMCPServers(agent AgentDefinition, platform Platform, bitriseToken string) map[string]mcpServerConfig {
	servers := make(map[string]mcpServerConfig)

	// Add agent's required MCPs for the detected platform.
	// Skip "bitrise" here — it requires a token and is handled below.
	if agent.RequiredMCPs != nil {
		for _, name := range agent.RequiredMCPs[platform] {
			if name == "bitrise" {
				continue
			}
			if preset, ok := mcpPresets[name]; ok {
				servers[name] = preset
			}
		}
	}

	// Add bitrise MCP only when token is available
	if bitriseToken != "" {
		preset := mcpPresets["bitrise"]
		preset.Headers = map[string]string{
			"Authorization": "Bearer " + bitriseToken,
		}
		servers["bitrise"] = preset
	}

	return servers
}

