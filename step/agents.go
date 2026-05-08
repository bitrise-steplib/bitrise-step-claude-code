package step

import (
	"embed"
	"encoding/json"
	"fmt"
	"strings"
)

//go:embed prompts/*.md
var promptFiles embed.FS

func loadPrompt(name string) string {
	data, err := promptFiles.ReadFile("prompts/" + name + ".md")
	if err != nil {
		panic(fmt.Sprintf("missing embedded prompt file for agent %q: %v", name, err))
	}
	return strings.TrimSpace(string(data))
}

// AgentDefinition is the JSON structure passed to --agents.
type AgentDefinition struct {
	Description  string                `json:"description"`
	Prompt       string                `json:"prompt"`
	Tools        string                `json:"tools,omitempty"`
	RequiredMCPs map[Platform][]string `json:"-"` // not serialized to CLI JSON
}

var builtinAgents = map[string]AgentDefinition{
	"code-review": {
		Description: "Reviews code changes for bugs, security issues, and style problems",
		Prompt:      loadPrompt("code-review"),
		Tools:       "Read,Grep,Glob",
		RequiredMCPs: map[Platform][]string{
			PlatformApple:   {"bitrise"},
			PlatformAndroid: {"bitrise"},
			PlatformOther:   {"bitrise"},
		},
	},
	"test-failure-analysis": {
		Description: "Analyzes test failures and suggests fixes",
		Prompt:      loadPrompt("test-failure-analysis"),
		Tools:       "Read,Grep,Glob,Bash",
		RequiredMCPs: map[Platform][]string{
			PlatformApple:   {"bitrise"},
			PlatformAndroid: {"bitrise"},
			PlatformOther:   {"bitrise"},
		},
	},
	"screenshot-generator": {
		Description: "Drives the app and captures App Store and Play Store screenshots",
		Prompt:      loadPrompt("screenshot-generator"),
		RequiredMCPs: map[Platform][]string{
			PlatformApple:   {"xcodebuild"},
			PlatformAndroid: {"mobile-mcp"},
			PlatformOther:   {"mobile-mcp"},
		},
	},
}

// getAgentDefinition returns the built-in agent definition and true, or zero value and false.
func getAgentDefinition(name string) (AgentDefinition, bool) {
	def, ok := builtinAgents[name]
	return def, ok
}

// resolveAgents resolves the agents input to a map suitable for --agents JSON.
// Accepts: "", "none", a built-in agent name, or raw JSON.
func resolveAgents(input string) (map[string]AgentDefinition, error) {
	if input == "" || input == "none" {
		return nil, nil
	}

	// Check if it's a built-in agent name
	if def, ok := builtinAgents[input]; ok {
		return map[string]AgentDefinition{input: def}, nil
	}

	// Try parsing as JSON
	if strings.HasPrefix(strings.TrimSpace(input), "{") {
		var agents map[string]AgentDefinition
		if err := json.Unmarshal([]byte(input), &agents); err != nil {
			return nil, fmt.Errorf("parse agents JSON: %w", err)
		}
		return agents, nil
	}

	return nil, fmt.Errorf("unknown agent %q: must be a built-in name or JSON object", input)
}
