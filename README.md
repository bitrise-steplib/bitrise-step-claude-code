# Claude Code

[![Step changelog](https://shields.io/github/v/release/bitrise-steplib/bitrise-step-claude-code?include_prereleases&label=changelog&color=blueviolet)](https://github.com/bitrise-steplib/bitrise-step-claude-code/releases)

Run Claude AI with a prompt and get the response

<details>
<summary>Description</summary>

Installs Claude Code CLI and runs it with the provided prompt.
Outputs the response as an environment variable and a file.

The step ships with built-in agents tailored for common mobile CI tasks.
Agents come with curated system prompts and automatically enable the MCP
servers they need based on the detected platform.
</details>

## Get started

Add this step directly to your workflow in the [Bitrise Workflow Editor](https://docs.bitrise.io/en/bitrise-ci/workflows-and-pipelines/steps/adding-steps-to-a-workflow.html).

You can also run this step directly with [Bitrise CLI](https://github.com/bitrise-io/bitrise).

## Built-in agents

Select an agent via the `agents` input. Each agent ships with a tailored system prompt and automatically configures the MCP servers it needs.

### `code-review`

Reviews code changes for bugs, security vulnerabilities, performance issues, and style problems. Includes platform-specific guidance for iOS/Swift, Android/Kotlin, React Native, Flutter/Dart, and Kotlin Multiplatform.

**Example prompt:**

```
Review the changes in this PR. Focus on security and correctness.
```

**Auto-enabled MCP servers:** Bitrise (when token is set)

### `test-failure-analysis`

Analyzes test failures and suggests concrete fixes. Looks for results in xcresult bundles, junit.xml, and Bitrise build logs.

**Example prompt:**

```
Analyze the test failures in this build and suggest fixes.
```

**Auto-enabled MCP servers:** Bitrise (when token is set)

### `screenshot-generator`

Drives the app on simulators/emulators, navigates through key screens, and captures screenshots for App Store and Play Store submission. References official store documentation for current size requirements.

**Example prompt:**

```
Generate App Store screenshots for my app. Key screens: onboarding, home, detail view, settings. Use light mode.
```

**Auto-enabled MCP servers:** XcodeBuildMCP (Apple projects) or Mobile MCP (Android/other)

### Custom agents

Pass a raw JSON object to define a custom agent:

```yaml
- agents: '{"my-agent":{"description":"My custom agent","prompt":"You are a helpful assistant"}}'
```

## Platform auto-detection

The step automatically detects your project platform to select the right MCP servers:

| Detected files | Platform | MCP server |
| --- | --- | --- |
| `.xcodeproj` / `.xcworkspace` | Apple | XcodeBuildMCP |
| `build.gradle` / `build.gradle.kts` | Android | Mobile MCP |
| Both Apple and Android markers | Other | Mobile MCP |
| Neither | Other | Mobile MCP |

Detection scans the working directory and one level of subdirectories (e.g. `ios/`, `android/`).

## Configuration

<details>
<summary>Inputs</summary>

| Key | Description | Flags | Default |
| --- | --- | --- | --- |
| `model` | The Claude model ID to use. Defaults to the Haiku model family for cost efficiency. See the full list of model IDs: https://platform.claude.com/docs/en/about-claude/models/overview | required | `claude-haiku-4-5` |
| `prompt` | The prompt to send to Claude Code. | required | |
| `api_key` | Your Anthropic API key. Defaults to the `ANTHROPIC_API_KEY` environment variable, so setting that as a Secret in Bitrise is the recommended approach. | required, sensitive | `$ANTHROPIC_API_KEY` |
| `allowed_tools` | Comma-separated list of tool names Claude is permitted to use. Maps to the `--allowed-tools` CLI flag. Example: `Read,Write,Edit,Bash` | | `Read,Write,Edit` |
| `agents` | Select a built-in agent (`code-review`, `test-failure-analysis`, `screenshot-generator`) or pass custom agent JSON. Use `none` for prompt-only mode. | | `none` |
| `bitrise_token` | Optional Bitrise PAT or WAT. When set, the step automatically configures the Bitrise MCP server so Claude can interact with Bitrise APIs. | sensitive | |
| `additional_cli_flags` | Any additional raw flags to pass to the `claude` command. These are appended after all other flags. Example: `--max-turns 5 --verbose` | | |
| `log_format` | `pretty` (default): Parses the real-time stream and formats it to resemble Claude Code's terminal UI. `raw`: Output is buffered and printed after Claude finishes. | required | `pretty` |
| `claude_version` | The version of Claude Code CLI to install. Use `latest` to always install the newest release, or pin to a specific version (e.g. `2.1.45`) for reproducibility. Also accepts `stable`. | required | `latest` |
</details>

<details>
<summary>Outputs</summary>

| Environment Variable | Description |
| --- | --- |
| `CLAUDE_OUTPUT` | The full text response from Claude. Truncated to 20KB if the response exceeds that limit. Use `CLAUDE_OUTPUT_FILE` for the full untruncated output. |
| `CLAUDE_OUTPUT_FILE` | Absolute path to a temporary file containing Claude's complete response, without any truncation. |
</details>

## Example workflows

### Code review on every PR

```yaml
workflows:
  code-review:
    steps:
    - git-clone: {}
    - claude-code:
        inputs:
        - prompt: "Review the changes in this PR. Focus on bugs, security, and code quality."
        - agents: code-review
        - api_key: $ANTHROPIC_API_KEY
        - bitrise_token: $BITRISE_TOKEN
```

### Analyze test failures

```yaml
workflows:
  test-analysis:
    steps:
    - git-clone: {}
    - xcode-test: {}
    - claude-code:
        is_always_run: true
        inputs:
        - prompt: "Analyze the test failures from the previous step and suggest fixes."
        - agents: test-failure-analysis
        - api_key: $ANTHROPIC_API_KEY
        - bitrise_token: $BITRISE_TOKEN
```

### Generate App Store screenshots

```yaml
workflows:
  screenshots:
    steps:
    - git-clone: {}
    - claude-code:
        inputs:
        - prompt: "Generate App Store screenshots. Key screens: onboarding, home, profile, settings."
        - agents: screenshot-generator
        - model: claude-sonnet-4-6
        - api_key: $ANTHROPIC_API_KEY
        - allowed_tools: Read,Write,Edit,Bash
```

## Contributing

We welcome [pull requests](https://github.com/bitrise-steplib/bitrise-step-claude-code/pulls) and [issues](https://github.com/bitrise-steplib/bitrise-step-claude-code/issues) against this repository.

For pull requests, work on your changes in a forked repository and use the Bitrise CLI to [run step tests locally](https://docs.bitrise.io/en/bitrise-ci/bitrise-cli/running-your-first-local-build-with-the-cli.html).

Learn more about developing steps:

- [Create your own step](https://docs.bitrise.io/en/bitrise-ci/workflows-and-pipelines/developing-your-own-bitrise-step/developing-a-new-step.html)
