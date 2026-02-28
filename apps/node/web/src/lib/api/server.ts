import "server-only";

const API_BACKEND_URL =
  process.env.API_BACKEND_URL ?? "http://localhost:8080";

export async function backendFetch(
  path: string,
  init?: RequestInit
): Promise<Response> {
  const url = `${API_BACKEND_URL}${path}`;
  return fetch(url, {
    ...init,
    headers: {
      "Content-Type": "application/json",
      ...init?.headers,
    },
  });
}
