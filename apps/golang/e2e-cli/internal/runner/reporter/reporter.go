package reporter

import (
	"fmt"
	"io"

	"github.com/user/micro-dp/e2e-cli/internal/runner"
)

func PrintConsole(w io.Writer, result runner.RunResult) {
	fmt.Fprintf(w, "E2E run: total=%d passed=%d failed=%d skipped=%d\n", result.Total, result.Passed, result.Failed, result.Skipped)
	for _, r := range result.Results {
		line := fmt.Sprintf("- [%s] %s (%s)", r.Status, r.ID, r.Duration)
		if r.Message != "" {
			line += ": " + r.Message
		}
		fmt.Fprintln(w, line)
	}
}

func WriteJSON(path string, result runner.RunResult) error {
	return runner.WriteJSON(path, result)
}
