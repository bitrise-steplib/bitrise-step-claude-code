package step

import "fmt"

const (
	claudeInstallURL = "https://claude.ai/install.sh"
	installRetries   = 3
)

func (s *Step) installClaude(version string) error {
	s.logger.Infof("Installing Claude Code CLI version %s...", version)

	// The bootstrap script accepts "latest", "stable", or a semver string as $1.
	// Verified against https://claude.ai/install.sh (redirects to bootstrap.sh on GCS):
	// TARGET="$1" is validated against ^(stable|latest|[0-9]+\.[0-9]+\.[0-9]+...) and
	// passed to the installer binary unchanged, so "latest" works natively here.
	installCmd := fmt.Sprintf("curl -fsSL %s | bash -s -- %s", claudeInstallURL, version)

	var lastErr error
	for attempt := 1; attempt <= installRetries; attempt++ {
		if attempt > 1 {
			s.logger.Warnf("Retrying install (attempt %d/%d)...", attempt, installRetries)
		}

		out, err := s.commandFactory.Create("bash", []string{"-c", installCmd}, nil).RunAndReturnTrimmedCombinedOutput()
		if err != nil {
			s.logger.Errorf("Install output:\n%s", out)
			lastErr = err
			continue
		}
		return nil
	}

	return fmt.Errorf("install failed after %d attempts: %w", installRetries, lastErr)
}
