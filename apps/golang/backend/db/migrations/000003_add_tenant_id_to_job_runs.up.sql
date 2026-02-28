INSERT INTO tenants (id, name) VALUES ('00000000-0000-0000-0000-000000000000', 'Default Tenant');

ALTER TABLE job_runs ADD COLUMN tenant_id TEXT REFERENCES tenants(id);

UPDATE job_runs SET tenant_id = '00000000-0000-0000-0000-000000000000' WHERE tenant_id IS NULL;

CREATE INDEX idx_job_runs_tenant_id ON job_runs(tenant_id);
