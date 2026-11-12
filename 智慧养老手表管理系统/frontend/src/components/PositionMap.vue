<template>
  <section class="position-map" aria-label="位置地图">
    <div class="position-map-heading">
      <div><h2>位置地图</h2><p>{{ valid ? `${selection?.name || '长者'} · ${selection?.lng}，${selection?.lat}` : '选择一条位置记录，在此查看。' }}</p></div>
      <el-button v-if="valid && mapPreferences.provider === 'amap' && !online" @click="enableOnlineMap">加载高德地图</el-button>
      <el-button v-if="online" link @click="disableOnlineMap">关闭在线地图</el-button>
      <a v-if="valid && mapPreferences.provider === 'tencent' && providerUrl"
        class="online-map-link" :href="providerUrl" target="_blank" rel="noopener noreferrer"
        referrerpolicy="no-referrer" @click="openInSystemMap">在腾讯地图查看位置</a>
      <a v-if="online && embeddedUrl" class="online-map-link" :href="embeddedUrl"
        target="_blank" rel="noopener noreferrer" referrerpolicy="no-referrer" @click="openInSystemMap">在高德地图中打开</a>
    </div>
    <iframe v-if="online && embeddedUrl" :src="embeddedUrl" class="position-map-canvas online-map"
      title="高德地图 · 设备最近上报位置" loading="lazy" referrerpolicy="strict-origin-when-cross-origin"
      allowfullscreen></iframe>
    <div v-show="!online" ref="canvas" class="position-map-canvas" :class="{ 'has-location': valid }" role="region" aria-label="离线位置标记"></div>
    <p v-if="valid" class="position-map-note">{{ describePositionContext(selection ?? {}) }}</p>
    <p class="position-map-note" role="status">{{ message }}</p>
  </section>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch, type PropType } from 'vue'
import L from 'leaflet'
import { isTauri, invoke } from '@tauri-apps/api/core'
import { ElMessage } from 'element-plus'
import 'leaflet/dist/leaflet.css'
import { describePositionContext, type PositionContext } from '../utils/positionContext'
import { mapPreferences, mapLocationUrl, mapDisplayPosition } from '../utils/mapProviders'

interface PositionSelection extends PositionContext {
  lng?: number | string | null
  lat?: number | string | null
  name?: string
}
const props = defineProps({ selection: { type: Object as PropType<PositionSelection | null>, default: null } })
const canvas = ref<HTMLElement | null>(null)
const online = ref(false)
const valid = computed(() => {
  const value = props.selection
  return value && value.lng != null && value.lat != null && String(value.lng).trim() !== '' && String(value.lat).trim() !== '' &&
    Number.isFinite(Number(value.lng)) && Number.isFinite(Number(value.lat)) && Math.abs(Number(value.lng)) <= 180 && Math.abs(Number(value.lat)) <= 90
})
const providerUrl = computed(() => {
  if (!valid.value || !props.selection || mapPreferences.provider === 'offline') return ''
  return mapLocationUrl(mapPreferences.provider, Number(props.selection.lng), Number(props.selection.lat), mapPreferences.style, mapPreferences.coordinateSystem)
})
const embeddedUrl = computed(() => mapPreferences.provider === 'amap' ? providerUrl.value : '')
const message = computed(() => {
  if (!valid.value) return '暂无可用坐标。'
  if (mapPreferences.provider === 'offline') return '当前只显示本地位置标记；未向地图服务发送坐标。'
  if (mapPreferences.provider === 'tencent') return '使用腾讯官方地图链接查看位置；仅在点击链接后发送坐标。'
  return online.value
    ? '高德底图使用转换后的 GCJ-02 显示坐标，仅供参考；原始上报数据未修改。'
    : '原始坐标已在本地标记。高德地图可按需加载，加载后将向服务商发送坐标。'
})
let map: L.Map | null = null
let marker: L.CircleMarker | null = null
let observer: ResizeObserver | null = null
let resizeFrame: number | null = null
function scheduleResize () {
  if (resizeFrame !== null) return
  resizeFrame = requestAnimationFrame(() => {
    resizeFrame = null
    map?.invalidateSize({ animate: false })
  })
}
function updateLocation () {
  if (!map) return
  if (marker) { marker.remove(); marker = null }
  if (!valid.value || !props.selection) { disableOnlineMap(); return }
  const point: L.LatLngTuple = [Number(props.selection.lat), Number(props.selection.lng)]
  map.setView(point, 15, { animate: false })
  marker = L.circleMarker(point, { radius: 9, color: '#fff', weight: 3, fillColor: '#215c4a', fillOpacity: 1 }).addTo(map)
  const label = document.createElement('span')
  label.textContent = props.selection.name || '最近上报位置'
  marker.bindTooltip(label, { direction: 'top', permanent: true, offset: [0, -12] })
}
async function openInSystemMap(event: MouseEvent) {
  if (!isTauri() || !props.selection || !valid.value || mapPreferences.provider === 'offline') return
  event.preventDefault()
  const [longitude, latitude] = mapDisplayPosition(Number(props.selection.lng), Number(props.selection.lat), mapPreferences.coordinateSystem)
  try {
    await invoke('open_map_location', { provider: mapPreferences.provider, latitude, longitude })
  } catch {
    ElMessage.error('无法打开系统浏览器，请检查桌面环境')
  }
}
function enableOnlineMap () {
  if (!map || !valid.value || mapPreferences.provider !== 'amap') return
  online.value = true
}
function disableOnlineMap () {
  online.value = false
  scheduleResize()
}
onMounted(() => {
  if (!canvas.value) return
  map = L.map(canvas.value, { scrollWheelZoom: false, attributionControl: true, zoomControl: false }).setView([30, 110], 5)
  L.control.zoom({ zoomInTitle: '放大地图', zoomOutTitle: '缩小地图' }).addTo(map)
  // Links are disabled in the desktop shell; retain readable attribution locally.
  map.attributionControl.setPrefix(false)
  observer = new ResizeObserver(scheduleResize)
  observer.observe(canvas.value)
  updateLocation()
  if (valid.value && mapPreferences.provider === 'amap' && mapPreferences.loadOnlineAutomatically) online.value = true
})
watch(() => props.selection, updateLocation)
watch(() => [mapPreferences.provider, mapPreferences.loadOnlineAutomatically, mapPreferences.coordinateSystem, props.selection?.lng, props.selection?.lat] as const, () => {
  online.value = mapPreferences.provider === 'amap' && mapPreferences.loadOnlineAutomatically && Boolean(valid.value)
  if (!online.value) scheduleResize()
})
onBeforeUnmount(() => { observer?.disconnect(); if (resizeFrame !== null) cancelAnimationFrame(resizeFrame); resizeFrame = null; disableOnlineMap(); map?.remove(); map = null })
</script>

<style scoped>
.position-map { margin-top: 24px; padding: clamp(16px, 3vw, 28px); border-top: 1px solid #dce3d8; }
.position-map-heading { display: flex; flex-wrap: wrap; align-items: center; justify-content: space-between; gap: 16px; margin-bottom: 18px; }
.position-map-heading h2 { margin: 0; font-size: 22px; font-weight: 500; }
.position-map-heading p, .position-map-note { color: var(--care-text-muted); line-height: 1.7; overflow-wrap: anywhere; }
.position-map-heading p { margin: 6px 0 0; }
.position-map-canvas { height: clamp(240px, 46vh, 440px); z-index: 0; border-radius: 16px; background-color: #eef1e9; background-image: linear-gradient(#dce3d833 1px, transparent 1px), linear-gradient(90deg, #dce3d833 1px, transparent 1px); background-size: 32px 32px; }
.online-map { display: block; width: 100%; border: 0; }
.online-map-link { color: var(--care-primary); font-size: 14px; text-underline-offset: 3px; }
.position-map-canvas:not(.has-location):not(.online-map) { opacity: .55; pointer-events: none; }
.position-map-note { margin: 12px 0 0; font-size: 15px; }
.position-map :deep(.leaflet-container) { font-family: inherit; }
.position-map :deep(.leaflet-control-zoom a) { width: 44px; height: 44px; line-height: 44px; }
.position-map :deep(.leaflet-control-attribution) { max-width: 100%; font-size: 12px; }
</style>
