import { describe, expect, it } from 'vitest'
import { formatCount, formatLatency, formatRate } from './format'

describe('formatCount', () => {
  it('passes small numbers through', () => {
    expect(formatCount(0)).toBe('0')
    expect(formatCount(999)).toBe('999')
  })

  it('scales to K/M/B with one decimal', () => {
    expect(formatCount(1_284)).toBe('1.3K')
    expect(formatCount(1_284_311)).toBe('1.3M')
    expect(formatCount(2_500_000_000)).toBe('2.5B')
  })

  it('drops the decimal at three digits', () => {
    expect(formatCount(128_431)).toBe('128K')
  })

  it('handles invalid input', () => {
    expect(formatCount(-1)).toBe('–')
    expect(formatCount(NaN)).toBe('–')
  })
})

describe('formatRate', () => {
  it('formats percentages', () => {
    expect(formatRate(0.0123)).toBe('1.23%')
    expect(formatRate(0)).toBe('0%')
  })

  it('floors tiny rates', () => {
    expect(formatRate(0.00005)).toBe('<0.01%')
  })

  it('handles invalid input', () => {
    expect(formatRate(-0.1)).toBe('–')
    expect(formatRate(Infinity)).toBe('–')
  })
})

describe('formatLatency', () => {
  it('uses ms below one second', () => {
    expect(formatLatency(87.4)).toBe('87ms')
  })

  it('promotes to seconds', () => {
    expect(formatLatency(1450)).toBe('1.45s')
  })

  it('handles invalid input', () => {
    expect(formatLatency(-5)).toBe('–')
  })
})
