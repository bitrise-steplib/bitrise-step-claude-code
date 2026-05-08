package step

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/bitrise-io/go-steputils/v2/stepconf"
	"github.com/bitrise-io/go-utils/v2/command"
	"github.com/bitrise-io/go-utils/v2/env"
	"github.com/bitrise-io/go-utils/v2/log"
)

type Input struct {
	Model              string `env:"model,required"`
	Prompt             string `env:"prompt,required"`
	APIKey             string `env:"api_key,required"`
	AllowedTools       string `env:"allowed_tools"`
	Agents             string `env:"agents"`
	BitriseToken       string `env:"bitrise_token"`
	AdditionalCLIFlags string `env:"additional_cli_flags"`
	ClaudeVersion      string `env:"claude_version,required"`
	LogFormat          string `env:"log_format,opt[pretty,raw]"`
}

type Step struct {
	logger         log.Logger
	inputParser    stepconf.InputParser
	commandFactory command.Factory
	envRepo        env.Repository
	exporter       outputExporter
}

func New(logger log.Logger, inputParser stepconf.InputParser, commandFactory command.Factory, envRepo env.Repository) *Step {
	return &Step{
		logger:         logger,
		inputParser:    inputParser,
		commandFactory: commandFactory,
		envRepo:        envRepo,
		exporter:       newExporter(commandFactory),
	}
}

func (s *Step) Run() error {
	var input Input
	if err := s.inputParser.Parse(&input); err != nil {
		return fmt.Errorf("parse inputs: %w", err)
	}

	stepconf.Print(input)

	if err := s.installClaude(input.ClaudeVersion); err != nil {
		return fmt.Errorf("install claude: %w", err)
	}

	// Add ~/.local/bin to PATH so the installed claude binary is found.
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("get home dir: %w", err)
	}
	localBin := filepath.Join(homeDir, ".local", "bin")
	currentPath := os.Getenv("PATH")
	if err := os.Setenv("PATH", localBin+string(os.PathListSeparator)+currentPath); err != nil {
		return fmt.Errorf("update PATH: %w", err)
	}

	// Resolve agent definition (built-in name, custom JSON, or none)
	agentDef, isBuiltin := getAgentDefinition(input.Agents)

	// Detect platform for MCP server selection
	workDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working dir: %w", err)
	}
	platform := detectPlatform(workDir)

	// Resolve MCP servers
	mcpServers := resolveMCPServers(agentDef, platform, input.BitriseToken)

	var mcpConfigPath string
	if len(mcpServers) > 0 {
		cfg := buildMCPConfig(mcpServers)
		mcpConfigPath, err = writeMCPConfigFile(cfg)
		if err != nil {
			return fmt.Errorf("write MCP config: %w", err)
		}
		defer os.Remove(mcpConfigPath)

		for name := range mcpServers {
			s.logger.Infof("MCP server enabled: %s", name)
		}
	}

	// Resolve agents input for CLI args
	var agentsJSON string
	if isBuiltin {
		agents := map[string]AgentDefinition{input.Agents: agentDef}
		data, err := json.Marshal(agents)
		if err != nil {
			return fmt.Errorf("marshal agent definition: %w", err)
		}
		agentsJSON = string(data)
		s.logger.Infof("Using built-in agent: %s", input.Agents)
	} else if input.Agents != "" && input.Agents != "none" {
		// Custom JSON pass-through
		agentsJSON = input.Agents
	}

	args, err := buildClaudeArgs(input, mcpConfigPath, agentsJSON)
	if err != nil {
		return fmt.Errorf("build claude args: %w", err)
	}

	output, err := s.runClaude(args, input)
	if err != nil {
		return fmt.Errorf("run claude: %w", err)
	}

	if err := s.exportOutputs(output); err != nil {
		return fmt.Errorf("export outputs: %w", err)
	}

	return nil
}
