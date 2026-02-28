export interface TrackerConfig {
  endpoint: string;
  tenantId?: string;
  userId?: string;
  anonymousId?: string;
  sessionId?: string;
  enabled?: boolean;
  debug?: boolean;
  flushIntervalMs?: number;
  maxQueueSize?: number;
  retryMaxAttempts?: number;
  retryBaseDelayMs?: number;
  headers?: Record<string, string>;
}

export interface TrackOptions {
  tenantId?: string;
  userId?: string;
  anonymousId?: string;
  sessionId?: string;
  eventTime?: string;
}

export interface TrackerEvent {
  event_id: string;
  tenant_id?: string;
  user_id?: string;
  anonymous_id?: string;
  session_id: string;
  event_name: string;
  properties: Record<string, unknown>;
  event_time: string;
  sent_at: string;
}

export interface FlushOptions {
  useBeacon?: boolean;
}
