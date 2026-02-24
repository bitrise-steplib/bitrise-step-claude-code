package step

import (
	"encoding/json"
	"fmt"
	"os"
)

// Claude MCP config JSON structure
type mcpConfig struct {
	McpServers map[string]mcpServer `json:"mcpServers"`
}

type mcpServer struct {
	Type    string            `json:"type"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers"`
}

func writeMCPConfig(token string) (string, error) {
	cfg := mcpConfig{
		McpServers: map[string]mcpServer{
			"bitrise": {
				Type: "sse",
				URL:  "https://mcp.bitrise.io",
				Headers: map[string]string{
					"Authorization": "Bearer " + token,
				},
			},
		},
	}

	data, err := json.Marshal(cfg)
	if err != nil {
		return "", fmt.Errorf("marshal MCP config: %w", err)
	}

	f, err := os.CreateTemp("", "bitrise-mcp-config-*.json")
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
