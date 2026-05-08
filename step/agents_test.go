package step

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAgentRegistry_AllBuiltinsExist(t *testing.T) {
	expected := []string{"code-review", "test-failure-analysis", "screenshot-generator"}
	for _, name := range expected {
		t.Run(name, func(t *testing.T) {
			_, ok := builtinAgents[name]
			assert.True(t, ok, "built-in agent %q should exist", name)
		})
	}
}

func TestAgentRegistry_DefinitionFields(t *testing.T) {
	for name, agent := range builtinAgents {
		t.Run(name, func(t *testing.T) {
			assert.NotEmpty(t, agent.Description, "Description")
			assert.NotEmpty(t, agent.Prompt, "Prompt")
		})
	}
}

func TestResolveAgents_BuiltinName(t *testing.T) {
	agents, err := resolveAgents("code-review")
	require.NoError(t, err)
	assert.Contains(t, agents, "code-review")
	assert.NotEmpty(t, agents["code-review"].Prompt)
}

func TestResolveAgents_None(t *testing.T) {
	agents, err := resolveAgents("none")
	require.NoError(t, err)
	assert.Nil(t, agents)
}

func TestResolveAgents_Empty(t *testing.T) {
	agents, err := resolveAgents("")
	require.NoError(t, err)
	assert.Nil(t, agents)
}

func TestResolveAgents_CustomJSON(t *testing.T) {
	customJSON := `{"my-agent":{"description":"Custom","prompt":"Do things"}}`

	agents, err := resolveAgents(customJSON)
	require.NoError(t, err)
	assert.Contains(t, agents, "my-agent")
	assert.Equal(t, "Custom", agents["my-agent"].Description)
}

func TestResolveAgents_InvalidJSON(t *testing.T) {
	_, err := resolveAgents("{invalid")
	require.Error(t, err)
}

func TestResolveAgents_UnknownName(t *testing.T) {
	_, err := resolveAgents("nonexistent-agent")
	require.Error(t, err)
}

func TestResolveAgents_JSONRoundtrip(t *testing.T) {
	agents, err := resolveAgents("code-review")
	require.NoError(t, err)

	data, err := json.Marshal(agents)
	require.NoError(t, err)

	var parsed map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(data, &parsed))
	assert.Contains(t, parsed, "code-review")
}

func TestAgentDefinition_ScreenshotGeneratorMCPs(t *testing.T) {
	agent := builtinAgents["screenshot-generator"]

	assert.Contains(t, agent.RequiredMCPs[PlatformApple], "xcodebuild")
	assert.Contains(t, agent.RequiredMCPs[PlatformAndroid], "mobile-mcp")
	assert.Contains(t, agent.RequiredMCPs[PlatformOther], "mobile-mcp")
}

func TestAgentDefinition_CodeReviewMCPs(t *testing.T) {
	agent := builtinAgents["code-review"]

	for _, p := range []Platform{PlatformApple, PlatformAndroid, PlatformOther} {
		mcps := agent.RequiredMCPs[p]
		assert.Contains(t, mcps, "bitrise", "platform %s", p)
	}

}


func TestGetAgentDefinition_Builtin(t *testing.T) {
	def, isBuiltin := getAgentDefinition("code-review")
	assert.True(t, isBuiltin)
	assert.NotEmpty(t, def.Prompt)
}

func TestGetAgentDefinition_NotBuiltin(t *testing.T) {
	_, isBuiltin := getAgentDefinition("custom-thing")
	assert.False(t, isBuiltin)
}

func TestGetAgentDefinition_None(t *testing.T) {
	_, isBuiltin := getAgentDefinition("none")
	assert.False(t, isBuiltin)
}
