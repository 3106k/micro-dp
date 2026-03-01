-- job_run_modules: 不足カラム追加
ALTER TABLE job_run_modules ADD COLUMN attempt INTEGER NOT NULL DEFAULT 1;
ALTER TABLE job_run_modules ADD COLUMN metrics_json TEXT;
ALTER TABLE job_run_modules ADD COLUMN error_code TEXT;

-- job_run_artifacts: 不足カラム追加
ALTER TABLE job_run_artifacts ADD COLUMN storage_type TEXT NOT NULL DEFAULT 'minio';
ALTER TABLE job_run_artifacts ADD COLUMN uri TEXT NOT NULL DEFAULT '';
ALTER TABLE job_run_artifacts ADD COLUMN checksum TEXT;
ALTER TABLE job_run_artifacts ADD COLUMN metadata_json TEXT;
