import { reactive } from 'vue'
import { wgs84ToGcj02 } from './gcj02'

export type MapProvider = 'amap' | 'tencent' | 'offline'
export type MapStyle = 'roadmap' | 'satellite'
export interface MapPreferences {
  provider: MapProvider
  style: MapStyle
  loadOnlineAutomatically: boolean
  reducedMotion: boolean
}
const defaults: MapPreferences = {
  provider: 'amap',
  style: 'roadmap',
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
export function mapLocationUrl(provider: Exclude<MapProvider, 'offline'>, longitude: number, latitude: number, style: MapStyle): string {
  const [lng, lat] = wgs84ToGcj02(longitude, latitude)
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
