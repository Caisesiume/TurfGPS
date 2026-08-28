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
  stubFetch(async () => {
    throw new Error('connection refused')
  })

  render(<App />)

  expect(await screen.findByText(/Unreachable — connection refused/)).toBeDefined()
})
