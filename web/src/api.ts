// The one call path to the service, per NFR-005's second criterion.

// Fixed into the bundle by Vite at build time; see .env.example for the
// default and why it is same-origin.
export const apiBaseUrl = import.meta.env.VITE_API_BASE_URL ?? '/api'

// fetchServiceGreeting calls the service's root route, which answers text/plain.
export async function fetchServiceGreeting(signal?: AbortSignal): Promise<string> {
  const response = await fetch(`${apiBaseUrl}/`, { signal })
  if (!response.ok) {
    throw new Error(`service responded ${response.status}`)
  }
  return (await response.text()).trim()
}
