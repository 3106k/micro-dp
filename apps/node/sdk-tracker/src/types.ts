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
  /** Write key for external (non-cookie) auth via collect API */
  writeKey?: string;
  /** Collect context data automatically (UA, referrer, screen, etc.) */
  collectContext?: boolean;
  /** Session inactivity timeout in ms (default: 1_800_000 = 30 min) */
  sessionTimeoutMs?: number;
}

export interface EventContext {
  user_agent?: string;
  referrer?: string;
  screen_width?: number;
  screen_height?: number;
  viewport_width?: number;
  viewport_height?: number;
  timezone?: string;
  page_url?: string;
  page_title?: string;
  language?: string;
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
  context?: EventContext;
}

export interface FlushOptions {
  useBeacon?: boolean;
}
