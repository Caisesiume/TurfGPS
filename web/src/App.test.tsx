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

test('calls the service at the absolute base URL fixed at build time', () => {
  const fetchSpy = stubFetch(async () => new Response('TurfGPS service\n'))

  render(<App />)

  expect(new URL(apiBaseUrl).protocol).toMatch(/^https?:$/)
  expect(fetchSpy.mock.calls[0]?.[0]).toBe(`${apiBaseUrl}/`)
})

test('renders what the service answered', async () => {
  stubFetch(async () => new Response('TurfGPS service\n'))

  render(<App />)

  expect(await screen.findByText('TurfGPS service')).toBeDefined()
})

test('says the service could not be reached when the call fails, and keeps the error as detail', async () => {
  // What a browser throws for a refused cross-origin answer and a dead network
  // alike: `Failed to fetch` in Chromium, `Load failed` in WebKit.
  stubFetch(async () => {
    throw new TypeError('Failed to fetch')
  })

  render(<App />)

  expect((await screen.findByRole('status')).textContent).toBe('The service could not be reached.')
  expect(screen.getByText('Failed to fetch')).toBeDefined()
})
