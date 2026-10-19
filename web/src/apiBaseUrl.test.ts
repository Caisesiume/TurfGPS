import { expect, test } from 'vitest'
import { LOCAL_SERVICE, parseApiBaseUrl, resolveApiBaseUrl } from './apiBaseUrl'

test('accepts an absolute http(s) URL and drops the trailing slash', () => {
  expect(parseApiBaseUrl('https://api.example.org/')).toBe('https://api.example.org')
  expect(parseApiBaseUrl(' http://127.0.0.1:8080 ')).toBe('http://127.0.0.1:8080')
  expect(parseApiBaseUrl('https://api.example.org/v1//')).toBe('https://api.example.org/v1')
})

test('refuses a missing or blank value, which names no service', () => {
  expect(() => parseApiBaseUrl(undefined)).toThrow(/is empty/)
  expect(() => parseApiBaseUrl('  ')).toThrow(/is empty/)
})

test('a Vercel build with the variable unset is refused, so a deploy cannot ship pointed at loopback', () => {
  expect(() => resolveApiBaseUrl(undefined, true)).toThrow(/unset on a Vercel build/)
  expect(() => resolveApiBaseUrl('  ', true)).toThrow(/unset on a Vercel build/)
})

test('any other build with the variable unset falls back to the local service', () => {
  expect(resolveApiBaseUrl(undefined, false)).toBe(LOCAL_SERVICE)
  expect(resolveApiBaseUrl('  ', false)).toBe(LOCAL_SERVICE)
})

test('a set value is validated on Vercel and off it', () => {
  expect(resolveApiBaseUrl('https://api.example.org/', true)).toBe('https://api.example.org')
  expect(resolveApiBaseUrl('https://api.example.org/', false)).toBe('https://api.example.org')
  expect(() => resolveApiBaseUrl('/api', true)).toThrow(/absolute URL/)
  expect(() => resolveApiBaseUrl('/api', false)).toThrow(/absolute URL/)
})

test('refuses a relative base, which only a rewrite could carry (ADR-0006 § D4)', () => {
  for (const relative of ['/api', '/', '//api.example.org', 'api.example.org']) {
    expect(() => parseApiBaseUrl(relative), relative).toThrow(/absolute URL/)
  }
})

test('refuses another scheme, credentials, a query or a fragment', () => {
  expect(() => parseApiBaseUrl('localhost:8080')).toThrow(/http or https/)
  expect(() => parseApiBaseUrl('ftp://api.example.org')).toThrow(/http or https/)
  expect(() => parseApiBaseUrl('https://user:pw@api.example.org')).toThrow(/no credentials/)
  expect(() => parseApiBaseUrl('https://api.example.org/?x=1')).toThrow(/no credentials/)
  expect(() => parseApiBaseUrl('https://api.example.org/#x')).toThrow(/no credentials/)
})
