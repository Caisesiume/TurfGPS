import { afterEach, expect, test, vi } from 'vitest'
import { cleanup, render, screen } from '@testing-library/react'
import App from './App'
import { apiBaseUrl } from './api'

afterEach(() => {
  cleanup()
  vi.unstubAllGlobals()
})

function stubFetch(implementation: typeof fetch) {
  const spy = vi.fn(implementation)
  vi.stubGlobal('fetch', spy)
  return spy
}

test('calls the service at the configured base URL', () => {
  const fetchSpy = stubFetch(async () => new Response('TurfGPS service\n'))

  render(<App />)

  expect(fetchSpy.mock.calls[0]?.[0]).toBe(`${apiBaseUrl}/`)
})

test('renders what the service answered', async () => {
  stubFetch(async () => new Response('TurfGPS service\n'))

  render(<App />)

  expect(await screen.findByText('TurfGPS service')).toBeDefined()
})

test('reports the service unreachable when the call fails', async () => {
  // What a browser actually throws when a fetch cannot complete: `Failed to
  // fetch` in Chromium, `Load failed` in WebKit. `connection refused` was a
  // stand-in that reads like copy somebody wrote, and asserting the rendered
  // text against it certified "whatever the network layer threw goes on the
  // screen" as correct user-facing behaviour. It is not — but the wording, and
  // a retry control, belong to the story that ships the planner UI and are
  // `future_work` rather than this item's, so what is asserted here is the
  // state the client reports and deliberately not the words it reports it in.
  stubFetch(async () => {
    throw new TypeError('Failed to fetch')
  })

  render(<App />)

  expect(await screen.findByText(/^Unreachable —/)).toBeDefined()
})
