package step

import (
	"fmt"
	"os"

	"github.com/bitrise-io/go-steputils/v2/export"
	"github.com/bitrise-io/go-utils/v2/command"
)

const maxOutputBytes = 20 * 1024 // 20KB

type outputExporter interface {
	ExportOutput(key, value string) error
	ExportOutputNoExpand(key, value string) error
}

func newExporter(commandFactory command.Factory) outputExporter {
	e := export.NewExporter(commandFactory)
	return &e
}

// exportOutputs writes the full output to a temp file and exports both
// CLAUDE_OUTPUT_FILE and CLAUDE_OUTPUT (truncated to 20KB) via envman.
func (s *Step) exportOutputs(output string) error {
	f, err := os.CreateTemp("", "claude-output-*.txt")
	if err != nil {
		return fmt.Errorf("create output temp file: %w", err)
	}
	defer f.Close()

	if _, err := f.WriteString(output); err != nil {
		return fmt.Errorf("write output file: %w", err)
	}

	if err := s.exporter.ExportOutput("CLAUDE_OUTPUT_FILE", f.Name()); err != nil {
		return err
	}

	truncated := output
	if len(output) > maxOutputBytes {
		s.logger.Warnf("Claude output exceeds 20KB (%d bytes); truncating CLAUDE_OUTPUT. Full output is in CLAUDE_OUTPUT_FILE.", len(output))
		truncated = output[:maxOutputBytes]
	}

	// --no-expand prevents envman from interpolating env var references in Claude's output.
	if err := s.exporter.ExportOutputNoExpand("CLAUDE_OUTPUT", truncated); err != nil {
		return err
	}

	s.logger.Infof("Output exported to %s", f.Name())
	return nil
}
