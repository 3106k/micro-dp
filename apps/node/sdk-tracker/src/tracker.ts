import { createId } from "./id";
import type {
  FlushOptions,
  TrackerConfig,
  TrackerEvent,
  TrackOptions
} from "./types";

const DEFAULTS = {
  enabled: true,
  debug: false,
  flushIntervalMs: 5000,
  maxQueueSize: 20,
  retryMaxAttempts: 3,
  retryBaseDelayMs: 300
};

class Tracker {
  private config: TrackerConfig = { endpoint: "", enabled: false };
  private queue: TrackerEvent[] = [];
  private timer: ReturnType<typeof setInterval> | null = null;
  private inFlight = false;
  private boundUnloadFlush = () => void this.flush({ useBeacon: true });

  init(config: TrackerConfig): void {
    this.config = {
      ...DEFAULTS,
      ...config,
      enabled: config.enabled ?? DEFAULTS.enabled
    };

    this.clearTimer();

    if (!this.config.enabled) {
      this.log("tracker disabled");
      return;
    }

    if (!this.config.endpoint) {
      this.config.enabled = false;
      this.log("tracker disabled because endpoint is empty");
      return;
    }

    this.timer = setInterval(() => {
      void this.flush();
    }, this.config.flushIntervalMs);

    if (typeof window !== "undefined") {
      window.addEventListener("visibilitychange", this.onVisibilityChange);
      window.addEventListener("pagehide", this.boundUnloadFlush);
      window.addEventListener("beforeunload", this.boundUnloadFlush);
    }
  }

  track(eventName: string, properties: Record<string, unknown> = {}, options: TrackOptions = {}): void {
    if (!this.config.enabled) {
      return;
    }

    const event = this.buildEvent(eventName, properties, options);
    this.queue.push(event);

    if (this.queue.length >= (this.config.maxQueueSize ?? DEFAULTS.maxQueueSize)) {
      void this.flush();
    }
  }

  identify(userId: string, traits: Record<string, unknown> = {}): void {
    this.track("identify", traits, { userId });
  }

  page(name = "page_view", properties: Record<string, unknown> = {}): void {
    this.track(name, properties);
  }

  async flush(options: FlushOptions = {}): Promise<void> {
    if (!this.config.enabled || this.inFlight || this.queue.length === 0) {
      return;
    }

    const batch = this.queue.splice(0, this.queue.length);
    this.inFlight = true;

    try {
      const beaconOk = options.useBeacon ? this.sendByBeacon(batch) : false;
      if (!beaconOk) {
        await this.sendWithRetry(batch);
      }
    } catch (error) {
      this.log("flush failed", error);
      this.queue.unshift(...batch);
    } finally {
      this.inFlight = false;
    }
  }

  private buildEvent(
    eventName: string,
    properties: Record<string, unknown>,
    options: TrackOptions
  ): TrackerEvent {
    const now = new Date().toISOString();

    return {
      event_id: createId(),
      tenant_id: options.tenantId ?? this.config.tenantId,
      user_id: options.userId ?? this.config.userId,
      anonymous_id: options.anonymousId ?? this.config.anonymousId,
      session_id: options.sessionId ?? this.config.sessionId ?? createId(),
      event_name: eventName,
      properties,
      event_time: options.eventTime ?? now,
      sent_at: now
    };
  }

  private sendByBeacon(events: TrackerEvent[]): boolean {
    if (typeof navigator === "undefined" || typeof navigator.sendBeacon !== "function") {
      return false;
    }

    try {
      const body = JSON.stringify({ events });
      const blob = new Blob([body], { type: "application/json" });
      const ok = navigator.sendBeacon(this.config.endpoint, blob);
      if (ok) {
        this.log("events sent by beacon", events.length);
      }
      return ok;
    } catch (error) {
      this.log("sendBeacon failed", error);
      return false;
    }
  }

  private async sendWithRetry(events: TrackerEvent[]): Promise<void> {
    const maxAttempts = this.config.retryMaxAttempts ?? DEFAULTS.retryMaxAttempts;
    const baseDelay = this.config.retryBaseDelayMs ?? DEFAULTS.retryBaseDelayMs;

    for (let attempt = 1; attempt <= maxAttempts; attempt += 1) {
      try {
        const response = await fetch(this.config.endpoint, {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
            ...(this.config.headers ?? {})
          },
          keepalive: true,
          body: JSON.stringify({ events })
        });

        if (!response.ok) {
          throw new Error(`HTTP ${response.status}`);
        }

        this.log("events sent", events.length);
        return;
      } catch (error) {
        if (attempt >= maxAttempts) {
          throw error;
        }

        const waitMs = baseDelay * 2 ** (attempt - 1);
        this.log(`retrying flush attempt=${attempt + 1} waitMs=${waitMs}`);
        await new Promise((resolve) => setTimeout(resolve, waitMs));
      }
    }
  }

  private onVisibilityChange = (): void => {
    if (typeof document !== "undefined" && document.visibilityState === "hidden") {
      void this.flush({ useBeacon: true });
    }
  };

  private clearTimer(): void {
    if (this.timer) {
      clearInterval(this.timer);
      this.timer = null;
    }

    if (typeof window !== "undefined") {
      window.removeEventListener("visibilitychange", this.onVisibilityChange);
      window.removeEventListener("pagehide", this.boundUnloadFlush);
      window.removeEventListener("beforeunload", this.boundUnloadFlush);
    }
  }

  private log(message: string, payload?: unknown): void {
    if (!this.config.debug) {
      return;
    }
    if (payload === undefined) {
      console.info(`[sdk-tracker] ${message}`);
      return;
    }
    console.info(`[sdk-tracker] ${message}`, payload);
  }
}

const tracker = new Tracker();

export function init(config: TrackerConfig): void {
  tracker.init(config);
}

export function track(
  eventName: string,
  properties: Record<string, unknown> = {},
  options: TrackOptions = {}
): void {
  tracker.track(eventName, properties, options);
}

export function identify(userId: string, traits: Record<string, unknown> = {}): void {
  tracker.identify(userId, traits);
}

export function page(name?: string, properties: Record<string, unknown> = {}): void {
  tracker.page(name, properties);
}

export async function flush(options: FlushOptions = {}): Promise<void> {
  await tracker.flush(options);
}
