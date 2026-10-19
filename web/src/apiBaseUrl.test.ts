import { expect, test } from 'vitest'
import { parseApiBaseUrl } from './apiBaseUrl'

test('accepts an absolute http(s) URL and drops the trailing slash', () => {
  expect(parseApiBaseUrl('https://api.example.org/')).toBe('https://api.example.org')
  expect(parseApiBaseUrl(' http://127.0.0.1:8080 ')).toBe('http://127.0.0.1:8080')
  expect(parseApiBaseUrl('https://api.example.org/v1//')).toBe('https://api.example.org/v1')
})

test('refuses a missing or blank value, so a build cannot ship without the service', () => {
  expect(() => parseApiBaseUrl(undefined)).toThrow(/is empty/)
  expect(() => parseApiBaseUrl('  ')).toThrow(/is empty/)
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
