# Claude Code

[![Step changelog](https://shields.io/github/v/release/bitrise-steplib/bitrise-step-claude-code?include_prereleases&label=changelog&color=blueviolet)](https://github.com/bitrise-steplib/bitrise-step-claude-code/releases)

Run Claude AI with a prompt and get the response

<details>
<summary>Description</summary>

Installs Claude Code CLI and runs it with the provided prompt.
Outputs the response as an environment variable and a file.
</details>

## 🧩 Get started

Add this step directly to your workflow in the [Bitrise Workflow Editor](https://docs.bitrise.io/en/bitrise-ci/workflows-and-pipelines/steps/adding-steps-to-a-workflow.html).

You can also run this step directly with [Bitrise CLI](https://github.com/bitrise-io/bitrise).

## ⚙️ Configuration

<details>
<summary>Inputs</summary>

| Key | Description | Flags | Default |
| --- | --- | --- | --- |
| `prompt` | The prompt to send to Claude Code. This is required. | required |  |
| `api_key` | Your Anthropic API key. Defaults to the `ANTHROPIC_API_KEY` environment variable, so setting that as a Secret in Bitrise is the recommended approach. | required, sensitive | `$ANTHROPIC_API_KEY` |
| `allowed_tools` | Comma-separated list of tool names Claude is permitted to use. Maps to the `--allowed-tools` CLI flag. Example: `Read,Write,Edit,Bash` |  | `Read,Write,Edit` |
| `agents` | Maps to the `--agents` CLI flag. Use `none` to use the default (prompt-only) mode. |  | `none` |
| `bitrise_token` | Optional Bitrise PAT or WAT. When set, the step automatically configures the Bitrise MCP server so Claude can interact with Bitrise APIs. | sensitive |  |
| `additional_cli_flags` | Any additional raw flags to pass to the `claude` command. These are appended after all other flags. Example: `--max-turns 5 --verbose` |  |  |
| `log_format` | `pretty` (default): Parses the real-time stream and formats it to resemble Claude Code's terminal UI. Intermediate steps and tool calls are shown in real time as they happen.  `raw`: Output is buffered and printed after Claude finishes, with no output parsing. Only the final response is printed. | required | `pretty` |
| `claude_version` | The version of Claude Code CLI to install. Use `latest` to always install the newest release, or pin to a specific version (e.g. `2.1.45`) for reproducibility. Also accepts `stable`. | required | `latest` |
</details>

<details>
<summary>Outputs</summary>

| Environment Variable | Description |
| --- | --- |
| `CLAUDE_OUTPUT` | The full text response from Claude. Truncated to 20KB if the response exceeds that limit. Use `CLAUDE_OUTPUT_FILE` for the full untruncated output. |
| `CLAUDE_OUTPUT_FILE` | Absolute path to a temporary file containing Claude's complete response, without any truncation. |
</details>

## 🙋 Contributing

We welcome [pull requests](https://github.com/bitrise-steplib/bitrise-step-claude-code/pulls) and [issues](https://github.com/bitrise-steplib/bitrise-step-claude-code/issues) against this repository.

For pull requests, work on your changes in a forked repository and use the Bitrise CLI to [run step tests locally](https://docs.bitrise.io/en/bitrise-ci/bitrise-cli/running-your-first-local-build-with-the-cli.html).

Learn more about developing steps:

- [Create your own step](https://docs.bitrise.io/en/bitrise-ci/workflows-and-pipelines/developing-your-own-bitrise-step/developing-a-new-step.html)
