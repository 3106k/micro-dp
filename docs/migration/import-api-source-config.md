# Import API: source_config Migration

## Summary

The Import Job creation API (`POST /api/v1/import/jobs`) has been migrated from
connector-specific top-level fields to a generic `source_config` object.

## What Changed

**Before (removed):**

```json
{
  "name": "my-import",
  "slug": "my-import",
  "connection_id": "conn-123",
  "spreadsheet_id": "abc123",
  "sheet_name": "Sheet1",
  "range": "A1:Z1000"
}
```

**After (current):**

```json
{
  "name": "my-import",
  "slug": "my-import",
  "connection_id": "conn-123",
  "source_config": {
    "spreadsheet_id": "abc123",
    "sheet_name": "Sheet1",
    "range": "A1:Z1000"
  }
}
```

## Key Points

- `source_config` is an optional JSON object. Connector-specific parameters that
  were previously top-level fields are now nested under this key.
- The old fields (`spreadsheet_id`, `sheet_name`, `range`) have been removed from
  the request schema. There is no backwards-compatibility shim.
- `connection_id` determines the connector type. `source_config` carries
  connector-specific configuration that varies by connector.
- If `source_config` is omitted, an empty config is used (valid for connectors
  that require no additional parameters).

## Migration Steps

1. Update request payloads to move connector-specific fields into `source_config`.
2. Update any client code or SDK integrations that construct `CreateImportJobRequest`.
3. No database migration is required — the change is API-layer only.
