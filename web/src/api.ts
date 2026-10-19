// The one call path to the service.
//
// The base is fixed into the bundle at build time by `vite.config.ts`, under
// `resolveApiBaseUrl`'s rule: a malformed value fails the build; an unset one
// fails a Vercel build and otherwise falls back to the local service. So the
// output stays static files and nothing is looked up at runtime.
export const apiBaseUrl: string = import.meta.env.VITE_API_BASE_URL

// fetchServiceGreeting calls the service's root route, which answers text/plain.
// A plain GET with no custom header is a CORS simple request, so the browser
// sends it without a preflight.
export async function fetchServiceGreeting(signal?: AbortSignal): Promise<string> {
  const response = await fetch(`${apiBaseUrl}/`, { signal })
  if (!response.ok) {
    throw new Error(`service responded ${response.status}`)
  }
  return (await response.text()).trim()
}
