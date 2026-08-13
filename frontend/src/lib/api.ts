// Thin fetch wrapper around the Go backend. Auth token attachment is a
// TODO once @auth0/nextjs-auth0 session wiring is finished — see the
// getAccessToken() TODO below.

const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL ?? 'http://localhost:8080/api/v1';

async function getAccessToken(): Promise<string | null> {
  // TODO: pull the Auth0 access token from the session, e.g. via
  // `getAccessToken()` from @auth0/nextjs-auth0 on the server, or a
  // client-side hook for client components.
  return null;
}

interface RequestOptions extends RequestInit {
  auth?: boolean;
}

export async function apiFetch<T>(path: string, options: RequestOptions = {}): Promise<T> {
  const { auth = true, headers, ...rest } = options;

  const finalHeaders: HeadersInit = {
    'Content-Type': 'application/json',
    ...headers,
  };

  if (auth) {
    const token = await getAccessToken();
    if (token) {
      (finalHeaders as Record<string, string>).Authorization = `Bearer ${token}`;
    }
  }

  const res = await fetch(`${API_BASE_URL}${path}`, { ...rest, headers: finalHeaders });

  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new Error(body.error ?? `Request failed with status ${res.status}`);
  }

  return res.json() as Promise<T>;
}

export const api = {
  listEvents: (params?: URLSearchParams) =>
    apiFetch(`/events${params ? `?${params}` : ''}`, { auth: false }),
  registerVolunteer: (payload: unknown) =>
    apiFetch('/volunteers/register', { method: 'POST', body: JSON.stringify(payload), auth: false }),
  myDashboard: () => apiFetch('/me/dashboard'),
  myVSR: () => apiFetch('/me/vsr'),
  submitServiceLog: (id: string) => apiFetch(`/service-logs/${id}/submit`, { method: 'POST' }),
};
