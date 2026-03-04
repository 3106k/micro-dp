package worker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/user/micro-dp/domain"
	"github.com/user/micro-dp/internal/observability"
	"github.com/user/micro-dp/usecase"
)

type JobRunConsumer struct {
	queue           domain.JobRunQueue
	jobRuns         domain.JobRunRepository
	transformWriter *TransformWriter
	sheetsWriter    *SheetsImportWriter
	credentials     *usecase.CredentialService
	connections     domain.ConnectionRepository
	metrics         *observability.JobRunMetrics
	metering        *usecase.MeteringService
}

func NewJobRunConsumer(
	queue domain.JobRunQueue,
	jobRuns domain.JobRunRepository,
	transformWriter *TransformWriter,
	sheetsWriter *SheetsImportWriter,
	credentials *usecase.CredentialService,
	connections domain.ConnectionRepository,
	metrics *observability.JobRunMetrics,
	metering *usecase.MeteringService,
) *JobRunConsumer {
	return &JobRunConsumer{
		queue:           queue,
		jobRuns:         jobRuns,
		transformWriter: transformWriter,
		sheetsWriter:    sheetsWriter,
		credentials:     credentials,
		connections:     connections,
		metrics:         metrics,
		metering:        metering,
	}
}

func (c *JobRunConsumer) Run(ctx context.Context) {
	log.Println("job_run_consumer started")

	for {
		select {
		case <-ctx.Done():
			log.Println("job_run_consumer stopped")
			return
		default:
			msg, err := c.queue.Dequeue(ctx)
			if err != nil {
				if ctx.Err() != nil {
					log.Println("job_run_consumer stopped")
					return
				}
				log.Printf("job_run_consumer: dequeue error: %v", err)
				continue
			}
			if msg == nil {
				continue
			}

			c.processMessage(ctx, msg)
		}
	}
}

func (c *JobRunConsumer) processMessage(ctx context.Context, msg *domain.JobRunMessage) {
	start := time.Now()

	// Idempotency check
	if err := c.queue.MarkProcessed(ctx, msg.JobRunID); err != nil {
		if errors.Is(err, domain.ErrJobRunAlreadyProcessed) {
			log.Printf("job_run_consumer: skipping duplicate job_run_id=%s", msg.JobRunID)
			c.metrics.DuplicateTotal.Add(ctx, 1)
			return
		}
		log.Printf("job_run_consumer: mark processed error job_run_id=%s: %v", msg.JobRunID, err)
		c.enqueueDLQ(ctx, msg, err.Error())
		return
	}

	// Fetch JobRun
	jr, err := c.jobRuns.FindByID(ctx, msg.TenantID, msg.JobRunID)
	if err != nil {
		log.Printf("job_run_consumer: find job run error job_run_id=%s: %v", msg.JobRunID, err)
		c.enqueueDLQ(ctx, msg, err.Error())
		return
	}

	// Parse RunSnapshot
	if jr.RunSnapshotJSON == nil || *jr.RunSnapshotJSON == "" {
		errMsg := "run_snapshot_json is empty"
		log.Printf("job_run_consumer: %s job_run_id=%s", errMsg, msg.JobRunID)
		c.failJobRun(ctx, msg, jr.ID, errMsg)
		return
	}

	var snapshot domain.RunSnapshot
	if err := json.Unmarshal([]byte(*jr.RunSnapshotJSON), &snapshot); err != nil {
		errMsg := fmt.Sprintf("unmarshal snapshot: %v", err)
		log.Printf("job_run_consumer: %s job_run_id=%s", errMsg, msg.JobRunID)
		c.failJobRun(ctx, msg, jr.ID, errMsg)
		return
	}

	// Update status to running
	if err := c.jobRuns.UpdateStarted(ctx, jr.ID); err != nil {
		log.Printf("job_run_consumer: update started error job_run_id=%s: %v", msg.JobRunID, err)
	}

	// Dispatch by job kind
	var execErr error
	switch snapshot.JobKind {
	case domain.JobKindTransform:
		execErr = c.executeTransform(ctx, msg, &snapshot)
	case domain.JobKindImport:
		execErr = c.executeImport(ctx, msg, &snapshot)
	default:
		execErr = fmt.Errorf("executor not implemented for kind: %s", snapshot.JobKind)
	}

	if execErr != nil {
		log.Printf("job_run_consumer: execute error job_run_id=%s: %v", msg.JobRunID, execErr)
		c.metrics.FailedTotal.Add(ctx, 1)
		c.failJobRun(ctx, msg, jr.ID, execErr.Error())
		c.metrics.Duration.Record(ctx, time.Since(start).Seconds())
		return
	}

	// Success
	if err := c.jobRuns.UpdateStatus(ctx, msg.TenantID, jr.ID, domain.StatusSuccess); err != nil {
		log.Printf("job_run_consumer: update success error job_run_id=%s: %v", msg.JobRunID, err)
	}

	c.metrics.ProcessedTotal.Add(ctx, 1)
	c.metrics.Duration.Record(ctx, time.Since(start).Seconds())

	log.Printf("job_run_consumer: completed job_run_id=%s kind=%s", msg.JobRunID, snapshot.JobKind)
}

func (c *JobRunConsumer) executeTransform(ctx context.Context, msg *domain.JobRunMessage, snapshot *domain.RunSnapshot) error {
	// Find transform module in snapshot
	var transformModule *domain.RunSnapshotModule
	for i := range snapshot.Modules {
		if snapshot.Modules[i].Category == domain.ModuleTypeCategoryTransform {
			transformModule = &snapshot.Modules[i]
			break
		}
	}
	if transformModule == nil {
		return fmt.Errorf("no transform module found in snapshot")
	}

	// Parse config_json to extract sql and dataset_ids
	var config struct {
		SQL        string   `json:"sql"`
		DatasetIDs []string `json:"dataset_ids"`
	}
	if err := json.Unmarshal([]byte(transformModule.ConfigJSON), &config); err != nil {
		return fmt.Errorf("parse transform config: %w", err)
	}

	if config.SQL == "" {
		return fmt.Errorf("transform config missing sql")
	}

	transformMsg := &domain.TransformJobMessage{
		JobRunID:   msg.JobRunID,
		TenantID:   msg.TenantID,
		SQL:        config.SQL,
		DatasetIDs: config.DatasetIDs,
		JobID:      snapshot.JobID,
		VersionID:  snapshot.VersionID,
	}

	result, err := c.transformWriter.Execute(ctx, transformMsg)
	if err != nil {
		return err
	}

	log.Printf("job_run_consumer: transform completed job_run_id=%s rows=%d output=%s",
		msg.JobRunID, result.RowCount, result.OutputKey)

	return nil
}

func (c *JobRunConsumer) failJobRun(ctx context.Context, msg *domain.JobRunMessage, jobRunID, errMsg string) {
	if err := c.jobRuns.UpdateFailed(ctx, jobRunID, errMsg); err != nil {
		log.Printf("job_run_consumer: update failed error job_run_id=%s: %v", msg.JobRunID, err)
	}
	c.enqueueDLQ(ctx, msg, errMsg)
}

func (c *JobRunConsumer) executeImport(ctx context.Context, msg *domain.JobRunMessage, snapshot *domain.RunSnapshot) error {
	// Find source module in snapshot
	var sourceModule *domain.RunSnapshotModule
	for i := range snapshot.Modules {
		if snapshot.Modules[i].Category == domain.ModuleTypeCategorySource {
			sourceModule = &snapshot.Modules[i]
			break
		}
	}
	if sourceModule == nil {
		return fmt.Errorf("no source module found in snapshot")
	}

	// Parse config_json
	var config struct {
		SpreadsheetID string `json:"spreadsheet_id"`
		SheetName     string `json:"sheet_name"`
		Range         string `json:"range"`
	}
	if err := json.Unmarshal([]byte(sourceModule.ConfigJSON), &config); err != nil {
		return fmt.Errorf("parse import config: %w", err)
	}
	if config.SpreadsheetID == "" {
		return fmt.Errorf("import config missing spreadsheet_id")
	}

	// Resolve connection → credential → access token
	if sourceModule.ConnectionID == nil {
		return fmt.Errorf("source module has no connection_id")
	}
	conn, err := c.connections.FindByID(ctx, msg.TenantID, *sourceModule.ConnectionID)
	if err != nil {
		return fmt.Errorf("find connection: %w", err)
	}
	if conn.CredentialID == nil {
		return fmt.Errorf("connection has no credential_id")
	}
	accessToken, err := c.credentials.GetValidAccessToken(ctx, msg.TenantID, *conn.CredentialID)
	if err != nil {
		return fmt.Errorf("get access token: %w", err)
	}

	importMsg := &SheetsImportMessage{
		JobRunID:      msg.JobRunID,
		TenantID:      msg.TenantID,
		SpreadsheetID: config.SpreadsheetID,
		SheetName:     config.SheetName,
		Range:         config.Range,
		AccessToken:   accessToken,
		JobID:         snapshot.JobID,
		VersionID:     snapshot.VersionID,
	}

	result, err := c.sheetsWriter.Execute(ctx, importMsg)
	if err != nil {
		return err
	}

	log.Printf("job_run_consumer: import completed job_run_id=%s rows=%d output=%s",
		msg.JobRunID, result.RowCount, result.OutputKey)

	return nil
}

func (c *JobRunConsumer) enqueueDLQ(ctx context.Context, msg *domain.JobRunMessage, reason string) {
	if err := c.queue.EnqueueDLQ(ctx, msg, reason); err != nil {
		log.Printf("job_run_consumer: enqueue dlq error job_run_id=%s: %v", msg.JobRunID, err)
	}
}
