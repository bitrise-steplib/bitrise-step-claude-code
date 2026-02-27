package step

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/bitrise-io/go-utils/v2/log/colorstring"
)

type streamEvent struct {
	Type         string         `json:"type"`
	Subtype      string         `json:"subtype"`
	Message      *streamMessage `json:"message"`
	Result       string         `json:"result"`
	IsError      bool           `json:"is_error"`
	NumTurns     int            `json:"num_turns"`
	DurationMS   int64          `json:"duration_ms"`
	TotalCostUSD float64        `json:"total_cost_usd"`
}

type streamMessage struct {
	Content []contentItem `json:"content"`
}

type contentItem struct {
	Type    string          `json:"type"`    // "text", "tool_use", "tool_result"
	Text    string          `json:"text"`
	Name    string          `json:"name"`    // tool_use: tool name
	Input   json.RawMessage `json:"input"`   // tool_use: arbitrary JSON
	IsError bool            `json:"is_error"`
	Content json.RawMessage `json:"content"` // tool_result: string or []contentItem
}

func parseStream(r io.Reader, logger log.Logger) (string, error) {
	scanner := bufio.NewScanner(r)
	buf := make([]byte, 1024*1024) // 1MB buffer — tool results can be very large
	scanner.Buffer(buf, len(buf))

	var finalOutput string
	var gotResult bool

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var event streamEvent
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			continue
		}

		if event.Type == "system" && event.Subtype == "init" {
			// Not interesting + don't warn about unrecognized event type.
			continue
		}

		switch event.Type {
		case "assistant":
			if event.Message == nil {
				continue
			}
			for _, item := range event.Message.Content {
				switch item.Type {
				case "text":
					logger.Printf("⏺ %s", item.Text)
				case "tool_use":
					detail := formatToolInput(item.Name, item.Input)
	
					if detail != "" {
						logger.Printf("%s %s(%s)", colorstring.Green("⏺"), item.Name, detail)
					} else {
						logger.Printf("%s %s", colorstring.Green("⏺"), item.Name)
					}
        default:
            logger.Printf("%s", item.Text)
				}
			}
		case "user":
			// tool_result items report whether each tool call succeeded or failed.
			// We print a second dot so the outcome is visible without having to scroll
			// back to find the original error in the tool's output.
			if event.Message == nil {
				continue
			}
			for _, item := range event.Message.Content {
				if item.Type != "tool_result" {
					continue
				}
				if item.IsError {
					logger.Printf("  ⎿  %s", colorstring.Red(toolResultText(item.Content)))
				} else {
					logger.Printf("  ⎿  %s", toolResultFirstLine(item.Content))
				}
			}
			logger.Println()
		case "result":
			gotResult = true
			if event.IsError {
				return "", fmt.Errorf("claude returned an error result: %s", strings.TrimSpace(event.Result))
			}
			durationSec := float64(event.DurationMS) / 1000.0
			logger.Println()
			logger.Printf("%s Completed in %.0fs (%d turns, $%.4f)",
				colorstring.Green("⏺"), durationSec, event.NumTurns, event.TotalCostUSD)
			finalOutput = strings.TrimSpace(event.Result)
    default:
      logger.Printf("Unrecognized event: %s", line)
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("read stream: %w", err)
	}

	if !gotResult {
		return "", fmt.Errorf("no result received from Claude")
	}

	return finalOutput, nil
}

// toolResultFirstLine and toolResultText extract content from a tool_result content field.
// The field can be either a plain JSON string or an array of content items (the API
// accepts both forms), so both functions try both.

func toolResultFirstLine(content json.RawMessage) string {
	s := toolResultText(content)
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			return line
		}
	}
	return ""
}

func toolResultText(content json.RawMessage) string {
	if len(content) == 0 {
		return ""
	}

	var s string
	if err := json.Unmarshal(content, &s); err == nil {
		return strings.TrimSpace(s)
	}

	var items []contentItem
	if err := json.Unmarshal(content, &items); err == nil {
		var parts []string
		for _, item := range items {
			if item.Type == "text" && strings.TrimSpace(item.Text) != "" {
				parts = append(parts, strings.TrimSpace(item.Text))
			}
		}
		return strings.Join(parts, "\n")
	}

	return ""
}

// formatToolInput picks the most useful single field from tool input JSON for display.
func formatToolInput(name string, input json.RawMessage) string {
	if len(input) == 0 {
		return ""
	}

	var fields map[string]any
	if err := json.Unmarshal(input, &fields); err != nil {
		return ""
	}

	var key string
	switch name {
	case "Bash":
		key = "command"
	case "Read", "Write", "Edit":
		key = "file_path"
	case "Glob", "Grep":
		key = "pattern"
	default:
		return ""
	}

	val, ok := fields[key]
	if !ok {
		return ""
	}

	str, ok := val.(string)
	if !ok {
		return ""
	}

	return str
}
