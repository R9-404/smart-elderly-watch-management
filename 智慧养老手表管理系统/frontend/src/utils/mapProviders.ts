import { reactive } from 'vue'
import { wgs84ToGcj02 } from './gcj02'

export type MapProvider = 'amap' | 'tencent' | 'offline'
export type MapStyle = 'roadmap' | 'satellite'
export type MapCoordinateSystem = 'wgs84' | 'gcj02'
export interface MapPreferences {
  provider: MapProvider
  style: MapStyle
  coordinateSystem: MapCoordinateSystem
  loadOnlineAutomatically: boolean
  reducedMotion: boolean
}
const defaults: MapPreferences = {
  provider: 'amap',
  style: 'roadmap',
  coordinateSystem: 'wgs84',
  loadOnlineAutomatically: false,
  reducedMotion: false
}
const storageKey = 'suining_map_preferences_v1'
function readPreferences(): MapPreferences {
  try {
    const data: unknown = JSON.parse(localStorage.getItem(storageKey) || 'null')
    if (!data || typeof data !== 'object') return { ...defaults }
    const value = data as Partial<MapPreferences>
    return {
      provider: value.provider === 'amap' || value.provider === 'tencent' || value.provider === 'offline' ? value.provider : defaults.provider,
      style: value.style === 'satellite' ? 'satellite' : 'roadmap',
      coordinateSystem: value.coordinateSystem === 'gcj02' ? 'gcj02' : 'wgs84',
      loadOnlineAutomatically: value.loadOnlineAutomatically === true,
      reducedMotion: value.reducedMotion === true
    }
  } catch { return { ...defaults } }
}
export const mapPreferences = reactive<MapPreferences>(readPreferences())
export function saveMapPreferences() {
  try { localStorage.setItem(storageKey, JSON.stringify(mapPreferences)) } catch { /* Session-only preferences. */ }
  document.documentElement.classList.toggle('reduce-motion', mapPreferences.reducedMotion)
}
export function resetMapPreferences() {
  Object.assign(mapPreferences, defaults)
  saveMapPreferences()
}
export function applyMapPreferences() {
  document.documentElement.classList.toggle('reduce-motion', mapPreferences.reducedMotion)
}
export function mapDisplayPosition(longitude: number, latitude: number, coordinateSystem: MapCoordinateSystem): [number, number] {
  return coordinateSystem === 'wgs84' ? wgs84ToGcj02(longitude, latitude) : [longitude, latitude]
}
export function mapLocationUrl(provider: Exclude<MapProvider, 'offline'>, longitude: number, latitude: number, style: MapStyle, coordinateSystem: MapCoordinateSystem = 'wgs84'): string {
  const [lng, lat] = mapDisplayPosition(longitude, latitude, coordinateSystem)
  if (provider === 'amap') {
    const params = new URLSearchParams({
      lng: String(lng), lat: String(lat), name: '设备最近上报位置',
      zoom: '16', maptype: style, platform: 'pc', src: 'suining'
    })
    return 'https://www.amap.com/ssr/embed/place?' + params
  }
  const params = new URLSearchParams({
    marker: `coord:${lat},${lng};title:设备最近上报位置`,
    referer: 'suining'
  })
  return 'https://apis.map.qq.com/uri/v1/marker?' + params
}
