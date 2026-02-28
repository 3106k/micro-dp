package runner

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/user/micro-dp/e2e-cli/internal/httpclient"
)

type Scenario interface {
	ID() string
	Run(ctx context.Context, client *httpclient.Client) error
}

type Status string

const (
	StatusPassed  Status = "passed"
	StatusFailed  Status = "failed"
	StatusSkipped Status = "skipped"
)

type ScenarioResult struct {
	ID       string `json:"id"`
	Status   Status `json:"status"`
	Duration string `json:"duration"`
	Message  string `json:"message,omitempty"`
}

type RunResult struct {
	Total     int              `json:"total"`
	Passed    int              `json:"passed"`
	Failed    int              `json:"failed"`
	Skipped   int              `json:"skipped"`
	StartedAt time.Time        `json:"started_at"`
	EndedAt   time.Time        `json:"ended_at"`
	Results   []ScenarioResult `json:"results"`
}

type Runner struct {
	client *httpclient.Client
}

func New(client *httpclient.Client) *Runner {
	return &Runner{client: client}
}

func (r *Runner) Run(ctx context.Context, scenarios []Scenario) RunResult {
	started := time.Now()
	result := RunResult{
		Total:     len(scenarios),
		StartedAt: started,
		Results:   make([]ScenarioResult, 0, len(scenarios)),
	}

	for _, s := range scenarios {
		begin := time.Now()
		err := s.Run(ctx, r.client)
		out := ScenarioResult{
			ID:       s.ID(),
			Duration: time.Since(begin).String(),
		}

		var skipped *SkipError
		switch {
		case err == nil:
			out.Status = StatusPassed
			result.Passed++
		case errors.As(err, &skipped):
			out.Status = StatusSkipped
			out.Message = skipped.Reason
			result.Skipped++
		default:
			out.Status = StatusFailed
			out.Message = err.Error()
			result.Failed++
		}

		result.Results = append(result.Results, out)
	}

	result.EndedAt = time.Now()
	return result
}

type SkipError struct {
	Reason string
}

func (e *SkipError) Error() string {
	return fmt.Sprintf("skipped: %s", e.Reason)
}

func Skip(reason string) error {
	return &SkipError{Reason: reason}
}

func WriteJSON(path string, result RunResult) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}
