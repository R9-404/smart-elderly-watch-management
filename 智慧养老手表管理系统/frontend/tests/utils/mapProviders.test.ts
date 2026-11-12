import { expect, it } from 'vitest'
import { mapLocationUrl, mapPreferences, resetMapPreferences, saveMapPreferences } from '@/utils/mapProviders'

it('builds only official map destinations with generic coordinate markers', () => {
  const amap = new URL(mapLocationUrl('amap', 116.397, 39.908, 'satellite'))
  expect(amap.origin).toBe('https://www.amap.com')
  expect(amap.pathname).toBe('/ssr/embed/place')
  expect(amap.searchParams.get('maptype')).toBe('satellite')
  expect(Number(amap.searchParams.get('lng'))).toBeGreaterThan(116.397)
  expect(amap.searchParams.get('name')).toBe('设备最近上报位置')
  const tencent = new URL(mapLocationUrl('tencent', 116.397, 39.908, 'roadmap'))
  expect(tencent.origin).toBe('https://apis.map.qq.com')
  expect(tencent.pathname).toBe('/uri/v1/marker')
  expect(tencent.searchParams.get('marker')).toContain('coord:')
  expect(tencent.toString()).not.toContain('长者姓名')
})

it('stores only supported map providers, display preferences and no credentials', () => {
  resetMapPreferences()
  mapPreferences.provider = 'tencent'
  mapPreferences.reducedMotion = true
  saveMapPreferences()
  expect(localStorage.getItem('suining_map_preferences_v1')).toContain('"provider":"tencent"')
  expect(document.documentElement.classList.contains('reduce-motion')).toBe(true)
  resetMapPreferences()
  expect(mapPreferences.provider).toBe('amap')
  expect(document.documentElement.classList.contains('reduce-motion')).toBe(false)
})
