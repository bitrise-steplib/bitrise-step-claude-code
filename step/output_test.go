package step

import (
	"os"
	"strings"
	"testing"

	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeExporter struct {
	exported   map[string]string
	noExpanded map[string]string
}

func newFakeExporter() *fakeExporter {
	return &fakeExporter{
		exported:   map[string]string{},
		noExpanded: map[string]string{},
	}
}

func (f *fakeExporter) ExportOutput(key, value string) error {
	f.exported[key] = value
	return nil
}

func (f *fakeExporter) ExportOutputNoExpand(key, value string) error {
	f.noExpanded[key] = value
	return nil
}

func newTestStep(exporter *fakeExporter) *Step {
	return &Step{
		logger:   log.NewLogger(),
		exporter: exporter,
	}
}

func TestExportOutputs_ExportsFilePath(t *testing.T) {
	exp := newFakeExporter()
	err := newTestStep(exp).exportOutputs("Hello from Claude")
	require.NoError(t, err)

	filePath, ok := exp.exported["CLAUDE_OUTPUT_FILE"]
	require.True(t, ok, "CLAUDE_OUTPUT_FILE not exported")
	assert.NotEmpty(t, filePath)
}

func TestExportOutputs_FileContainsFullOutput(t *testing.T) {
	exp := newFakeExporter()
	output := "Full Claude response here"

	err := newTestStep(exp).exportOutputs(output)
	require.NoError(t, err)

	content, err := os.ReadFile(exp.exported["CLAUDE_OUTPUT_FILE"])
	require.NoError(t, err)
	assert.Equal(t, output, string(content))
}

func TestExportOutputs_OutputUnderLimit(t *testing.T) {
	exp := newFakeExporter()
	output := "Short response"

	err := newTestStep(exp).exportOutputs(output)
	require.NoError(t, err)

	assert.Equal(t, output, exp.noExpanded["CLAUDE_OUTPUT"])
}

func TestExportOutputs_TruncatesLargeOutput(t *testing.T) {
	exp := newFakeExporter()
	large := strings.Repeat("a", maxOutputBytes+500)

	err := newTestStep(exp).exportOutputs(large)
	require.NoError(t, err)

	assert.Equal(t, maxOutputBytes, len(exp.noExpanded["CLAUDE_OUTPUT"]))

	// The file always gets the full untruncated output.
	content, err := os.ReadFile(exp.exported["CLAUDE_OUTPUT_FILE"])
	require.NoError(t, err)
	assert.Equal(t, len(large), len(string(content)))
}
