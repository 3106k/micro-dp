export async function readApiErrorMessage(
  response: Response,
  fallback = "Request failed"
): Promise<string> {
  try {
    const data = (await response.clone().json()) as {
      error?: unknown;
      message?: unknown;
    };
    if (typeof data.error === "string" && data.error.length > 0) {
      return data.error;
    }
    if (typeof data.message === "string" && data.message.length > 0) {
      return data.message;
    }
    return fallback;
  } catch {
    return fallback;
  }
}

export function toErrorMessage(
  error: unknown,
  fallback = "Unexpected error"
): string {
  if (error instanceof Error && error.message.length > 0) {
    return error.message;
  }
  return fallback;
}
