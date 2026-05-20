<template>
  <div ref="mapEl" class="app-map" :style="{ height }"></div>
</template>

<script setup lang="ts">
import { onMounted, onBeforeUnmount, ref, watch } from 'vue'
import L from 'leaflet'
import 'leaflet/dist/leaflet.css'

type MarkerItem = {
  lat: number
  lng: number
  title?: string
  subtitle?: string
}

const props = withDefaults(
  defineProps<{
    center?: { lat: number; lng: number } | null
    markers?: MarkerItem[]
    zoom?: number
    height?: string
  }>(),
  {
    center: null,
    markers: () => [],
    zoom: 13,
    height: '320px',
  },
)

const mapEl = ref<HTMLElement | null>(null)

let map: L.Map | null = null
let markersLayer: L.LayerGroup | null = null

delete (L.Icon.Default.prototype as any)._getIconUrl
L.Icon.Default.mergeOptions({
  iconRetinaUrl: 'https://unpkg.com/leaflet@1.9.4/dist/images/marker-icon-2x.png',
  iconUrl: 'https://unpkg.com/leaflet@1.9.4/dist/images/marker-icon.png',
  shadowUrl: 'https://unpkg.com/leaflet@1.9.4/dist/images/marker-shadow.png',
})

function validMarkers() {
  return props.markers.filter(
    (m) =>
      Number.isFinite(m.lat) &&
      Number.isFinite(m.lng) &&
      !(m.lat === 0 && m.lng === 0),
  )
}

function renderMap() {
  if (!map) return

  const items = validMarkers()

  if (markersLayer) {
    markersLayer.clearLayers()
  } else {
    markersLayer = L.layerGroup().addTo(map)
  }

  items.forEach((item) => {
  // 1. Создаём маркер
  const marker = L.marker([item.lat, item.lng])

  const popupHtml = [item.title, item.subtitle]
    .filter(Boolean)
    .join('<br>')

  if (popupHtml) {
    marker.bindPopup(popupHtml)
  }

  marker.addTo(markersLayer!)

  // 2. Рисуем область работ (окружность 100 метров)
  const workArea = L.circle([item.lat, item.lng], {
    radius: 100,              // радиус в метрах
    color: '#4f46e5',         // цвет границы
    fillColor: '#4f46e5',     // цвет заливки
    fillOpacity: 0.15,        // прозрачность заливки
    weight: 2,                // толщина границы
    interactive: false        // окружность не ловит клики (чтобы не мешать карте)
  })
  workArea.addTo(markersLayer!)
})

  if (items.length > 1) {
    const bounds = L.latLngBounds(
      items.map((item) => [item.lat, item.lng] as [number, number]),
    )
    map.fitBounds(bounds, { padding: [24, 24] })
    return
  }

  if (items.length === 1) {
    const first = items[0]
    if (!first) return
    map.setView([first.lat, first.lng], props.zoom)
    return
  }

  if (props.center) {
    map.setView([props.center.lat, props.center.lng], props.zoom)
  } else {
    map.setView([64.5393, 40.5187], 11)
  }
}

onMounted(() => {
  if (!mapEl.value) return

  map = L.map(mapEl.value, {
    zoomControl: true,
    attributionControl: false,
  })

  L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
    maxZoom: 19,
  }).addTo(map)

  renderMap()

  setTimeout(() => {
    map?.invalidateSize()
  }, 0)
})

watch(
  () => [props.center, props.markers],
  () => {
    renderMap()
    setTimeout(() => {
      map?.invalidateSize()
    }, 0)
  },
  { deep: true },
)

onBeforeUnmount(() => {
  map?.remove()
  map = null
})
</script>

<style scoped>
.app-map {
  width: 100%;
  border-radius: 12px;
  overflow: hidden;
  border: 1px solid #e5e7eb;
  background: #f3f4f6;
}
</style>