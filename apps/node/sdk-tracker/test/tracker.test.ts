import { beforeEach, describe, expect, it, vi } from "vitest";

import { flush, init, track } from "../src/index";

describe("sdk-tracker", () => {
  beforeEach(() => {
    vi.restoreAllMocks();
    vi.useFakeTimers();
    // @ts-expect-error test stub
    global.fetch = vi.fn().mockResolvedValue({ ok: true, status: 200 });
    Object.defineProperty(global, "navigator", {
      value: {
        sendBeacon: vi.fn().mockReturnValue(true)
      },
      configurable: true
    });
  });

  it("sends events with fetch on flush", async () => {
    init({ endpoint: "http://localhost:8080/api/v1/events", enabled: true });
    track("signup_clicked", { source: "hero" });

    await flush();

    expect(fetch).toHaveBeenCalledTimes(1);
    const [, initArg] = vi.mocked(fetch).mock.calls[0];
    const body = JSON.parse((initArg as RequestInit).body as string);
    expect(body.events[0].event_name).toBe("signup_clicked");
    expect(body.events[0].event_id).toBeTruthy();
  });

  it("uses sendBeacon when requested", async () => {
    init({ endpoint: "http://localhost:8080/api/v1/events", enabled: true });
    track("page_view", { path: "/" });

    await flush({ useBeacon: true });

    expect(navigator.sendBeacon).toHaveBeenCalledTimes(1);
  });

  it("retries with exponential backoff", async () => {
    const fetchMock = vi
      .fn()
      .mockRejectedValueOnce(new Error("network"))
      .mockRejectedValueOnce(new Error("network"))
      .mockResolvedValue({ ok: true, status: 200 });
    // @ts-expect-error test stub
    global.fetch = fetchMock;

    init({
      endpoint: "http://localhost:8080/api/v1/events",
      enabled: true,
      retryMaxAttempts: 3,
      retryBaseDelayMs: 10
    });

    track("purchase", { amount: 1000 });
    const promise = flush();
    await vi.advanceTimersByTimeAsync(100);
    await promise;

    expect(fetchMock).toHaveBeenCalledTimes(3);
  });
});
