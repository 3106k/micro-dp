import "server-only";

const API_BACKEND_URL =
  process.env.API_BACKEND_URL ?? "http://localhost:8080";

export async function backendFetch(
  path: string,
  init?: RequestInit
): Promise<Response> {
  const url = `${API_BACKEND_URL}${path}`;
  const headers = new Headers(init?.headers);
  if (init?.body && !headers.has("Content-Type")) {
    headers.set("Content-Type", "application/json");
  }
  if (!headers.has("Accept")) {
    headers.set("Accept", "application/json");
  }

  return fetch(url, {
    ...init,
    headers,
  });
}
