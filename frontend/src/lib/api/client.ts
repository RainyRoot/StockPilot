const BASE_URL = 'http://localhost:8080/api/v1';

interface ApiResponse<T> {
  data: T;
  error: { code: string; message: string } | null;
  meta: { timestamp: number };
}

export async function api<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(`${BASE_URL}${path}`, {
    headers: {
      'Content-Type': 'application/json',
      ...options?.headers,
    },
    ...options,
  });

  const json: ApiResponse<T> = await res.json();

  if (json.error) {
    throw new Error(json.error.message);
  }

  return json.data;
}
