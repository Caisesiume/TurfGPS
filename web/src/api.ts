// The one call path to the service, per NFR-005's second criterion.

// Fixed into the bundle by Vite at build time; see .env.example for the
// default and why it is same-origin.
//
// An unset variable arrives as undefined, but one set to nothing arrives as the
// empty string, which ?? passes straight through. That value builds a request
// for `/` — the client's own index.html, answered 200 with a document body — so
// an unconfigured build reports its own markup as the service's greeting and
// reads as reachable. A blank value is therefore treated as absent: it resolves
// to the documented default, which App renders, so the base the client actually
// used stays on screen rather than being inferred from a passing check.
const configuredBaseUrl = import.meta.env.VITE_API_BASE_URL?.trim() ?? ''
export const apiBaseUrl = configuredBaseUrl === '' ? '/api' : configuredBaseUrl

// fetchServiceGreeting calls the service's root route, which answers text/plain.
export async function fetchServiceGreeting(signal?: AbortSignal): Promise<string> {
  const response = await fetch(`${apiBaseUrl}/`, { signal })
  if (!response.ok) {
    throw new Error(`service responded ${response.status}`)
  }
  return (await response.text()).trim()
}
