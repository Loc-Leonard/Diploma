<template>
  <div class="customer-layout">
    <CustomerLayout />

    <main class="customer-main">
      <header class="customer-header">
        <h1 class="customer-title">Новый объект</h1>
      </header>

      <!-- одна колонка, форма по центру -->
      <section class="create-object-wrapper">
        <div class="column column--objects">
          <div class="column-header">
            <h2>Данные объекта</h2>
          </div>

          <form class="object-form" @submit.prevent="submit">
            <div class="form-grid">
              <div class="form-field">
                <label>Название объекта</label>
                <input v-model="form.name" required />
              </div>

              <!-- ГОРОД -->
              <div class="form-field">
                <label>Город</label>
                <input 
                  v-model="form.city"
                  required
                  @focus="onCityFocus"
                  @blur="onCityBlur"
                  placeholder="Начните вводить город..."
                  autocomplete="off"
                />
                <!-- Выпадающий список -->
                <ul
                  v-if="showCityDropdown"
                  class="autocomplete-list"
                  @mouseenter="cancelBlur(cityBlurTimer)"
                  @mouseleave="cancelBlur(cityBlurTimer)"
                >
                  <li v-if="cityLoading" class="autocomplete-status">⏳ Загрузка...</li>
                  <li v-else-if="citySuggestions.length === 0 && form.city?.length >= 2" class="autocomplete-status">Ничего не найдено</li>
                  <li
                    v-else
                    v-for="item in citySuggestions"
                    :key="item.label"
                    @mousedown.prevent="selectCity(item)"
                  >{{ item.label }}</li>
                </ul>
              </div>

              <!-- АДРЕС -->
              <div class="form-field form-field--full">
                <label>Адрес</label>
                <input 
                  v-model="form.address"
                  required 
                  @focus="onAddressFocus"
                  @blur="onAddressBlur"
                  placeholder="Например: ул. Ленина, 10"
                  autocomplete="off"
                />
                <!-- Выпадающий список -->
                <ul
                  v-if="showAddressDropdown"
                  class="autocomplete-list"
                  @mouseenter="cancelBlur(addressBlurTimer)"
                  @mouseleave="cancelBlur(addressBlurTimer)"
                >
                  <li v-if="addressLoading" class="autocomplete-status">⏳ Загрузка...</li>
                  <li v-else-if="addressSuggestions.length === 0 && form.address?.length >= 3" class="autocomplete-status">Ничего не найдено</li>
                  <li
                    v-else
                    v-for="item in addressSuggestions"
                    :key="item.label"
                    @mousedown.prevent="selectAddress(item)"
                  >{{ item.label }}</li>
                </ul>
              </div>

              <div class="form-field form-field--full">
                <label>Описание</label>
                <textarea v-model="form.description" rows="3" />
              </div>

              <div class="form-field">
                <label>Плановая дата начала</label>
                <input type="date" v-model="form.plannedStartDate" />
              </div>

              <div class="form-field">
                <label>Плановая дата окончания</label>
                <input type="date" v-model="form.plannedEndDate" />
              </div>

              <div class="form-field">
                <label>Прораб</label>
                <select v-model.number="form.foremanUserId" required>
                  <option value="" disabled>Выберите прораба</option>
                  <option
                    v-for="u in foremen"
                    :key="u.id"
                    :value="u.id"
                  >
                    {{ u.full_name }}
                  </option>
                </select>
              </div>

              <div class="form-field">
                <label>Инспектор</label>
                <select v-model.number="form.inspectorUserId" required>
                  <option value="" disabled>Выберите инспектора</option>
                  <option
                    v-for="u in inspectors"
                    :key="u.id"
                    :value="u.id"
                  >
                    {{ u.full_name }}
                  </option>
                </select>
              </div>
            </div>

            <div class="form-actions">
              <button
                type="button"
                class="secondary-btn"
                @click="goToObjects"
              >
                Отмена
              </button>
              <button
                type="submit"
                class="primary-btn"
                :disabled="submitting"
              >
                Создать объект
              </button>
            </div>
          </form>
        </div>
      </section>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch, type Ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import CustomerLayout from './CustomerLayout.vue'

const API_BASE = 'http://localhost:8080'
const router = useRouter()
const auth = useAuthStore()

type SimpleUser = {
  id: number
  full_name: string
}

// Интерфейс для элемента подсказки (с координатами)
interface SuggestionItem {
  label: string
  lat: string
  lng: string
}

const foremen = ref<SimpleUser[]>([])
const inspectors = ref<SimpleUser[]>([])
const submitting = ref(false)

const form = ref({
  name: '',
  city: '',
  address: '',
  description: '',
  plannedStartDate: '',
  plannedEndDate: '',
  foremanUserId: 0,
  inspectorUserId: 0,
  // Добавляем поля для координат
  lat: 0,
  lng: 0,
})

// ========== АВТОКОМПЛИТ ДЛЯ ГОРОДА И АДРЕСА ==========

// Храним объекты с координатами, а не просто строки
const citySuggestions = ref<SuggestionItem[]>([])
const addressSuggestions = ref<SuggestionItem[]>([])

const showCityDropdown = ref(false)
const showAddressDropdown = ref(false)
const cityLoading = ref(false)
const addressLoading = ref(false)

const cityBlurTimer = ref<number | null>(null)
const addressBlurTimer = ref<number | null>(null)

// AbortController для отмены зависших запросов
let cityAbortCtrl: AbortController | null = null
let addressAbortCtrl: AbortController | null = null

// 🔹 Утилита debounce — ждёт 300мс после последнего ввода
function debounce<T extends (...args: any[]) => void>(func: T, wait: number) {
  let timeout: number | null = null
  return (...args: Parameters<T>) => {
    if (timeout !== null) clearTimeout(timeout)
    timeout = window.setTimeout(() => func(...args), wait)
  }
}

// 🔹 Запрос подсказок городов (Nominatim API)
async function fetchCitySuggestions(query: string) {
  cityAbortCtrl?.abort()
  cityAbortCtrl = new AbortController()

  if (!query || query.length < 2) {
    citySuggestions.value = []
    return
  }

  cityLoading.value = true
  try {
    const url = `https://nominatim.openstreetmap.org/search?q=${encodeURIComponent(query + ', Россия')}&format=json&limit=5&countrycodes=ru&addressdetails=1`
    
    const res = await fetch(url, { 
      signal: cityAbortCtrl.signal,
      headers: { 'User-Agent': 'ConstructionApp/1.0 (your@email.com)' } 
    })

    if (!res.ok) throw new Error(`HTTP ${res.status}`)
    const data = await res.json()

    // Сохраняем объекты с координатами и названием
    citySuggestions.value = data
      .map((item: any) => ({
        label: item.address?.city || item.address?.town || item.address?.village || item.display_name.split(',')[0],
        lat: item.lat,
        lng: item.lon
      }))
      .filter(Boolean)
      // Убираем дубликаты по названию
      .filter((v: SuggestionItem, i: number, a: SuggestionItem[]) => a.findIndex(x => x.label === v.label) === i)

  } catch (e: any) {
    if (e.name !== 'AbortError') console.error('City API error:', e)
    citySuggestions.value = []
  } finally {
    cityLoading.value = false
  }
}

// 🔹 Запрос подсказок адресов
async function fetchAddressSuggestions(query: string) {
  addressAbortCtrl?.abort()
  addressAbortCtrl = new AbortController()

  if (!query || query.length < 3) {
    addressSuggestions.value = []
    return
  }

  addressLoading.value = true
  try {
    // 1. Улучшаем запрос: добавляем город из формы, если он заполнен
    // Это поможет найти "Новгородский пр." именно в Архангельске, а не в Великом Новгороде
    const cityQuery = form.value.city ? `, ${form.value.city}` : ''
    
    const url = `https://nominatim.openstreetmap.org/search?q=${encodeURIComponent(query + cityQuery + ', Россия')}&format=json&limit=5&countrycodes=ru&addressdetails=1`
    
    const res = await fetch(url, { 
      signal: addressAbortCtrl.signal,
      headers: { 'User-Agent': 'ConstructionApp/1.0 (your@email.com)' } 
    })

    if (!res.ok) throw new Error(`HTTP ${res.status}`)
    const data = await res.json()

    // 2. Формируем красивый краткий адрес "Улица, Дом"
    addressSuggestions.value = data.map((item: any) => {
      const addr = item.address || {}
      
      // Пытаемся найти название дороги и номер дома в детальной структуре ответа
      const road = addr.road || addr.footway || addr.pedestrian || ''
      const house = addr.house_number || ''
      
      // По умолчанию берем полный адрес (на случай, если структура неизвестна)
      let label = item.display_name

      // Если удалось найти улицу и дом, собираем их в короткую строку
      if (road) {
        label = road
        if (house) label += `, ${house}`
      }
      
      return {
        label,       // Короткий адрес для отображения
        lat: item.lat,
        lng: item.lon
      }
    })
  } catch (e: any) {
    if (e.name !== 'AbortError') console.error('Address API error:', e)
    addressSuggestions.value = []
  } finally {
    addressLoading.value = false
  }
}

// 🔹 Debounced-версии функций (задержка 300мс)
const debouncedFetchCity = debounce(fetchCitySuggestions, 300)
const debouncedFetchAddress = debounce(fetchAddressSuggestions, 300)

// 🔹 Следим за вводом пользователя
watch(() => form.value.city, (val) => debouncedFetchCity(val || ''))
watch(() => form.value.address, (val) => debouncedFetchAddress(val || ''))

// 🔹 Обработчики выбора подсказки
function selectCity(item: SuggestionItem) {
  form.value.city = item.label
  form.value.lat = parseFloat(item.lat) // Сохраняем координаты!
  form.value.lng = parseFloat(item.lng)
  showCityDropdown.value = false
  citySuggestions.value = []
}

function selectAddress(item: SuggestionItem) {
  form.value.address = item.label
  form.value.lat = parseFloat(item.lat) // Сохраняем координаты (адрес точнее города)!
  form.value.lng = parseFloat(item.lng)
  showAddressDropdown.value = false
  addressSuggestions.value = []
}

// 🔹 Управление фокусом/потерей фокуса
function onCityFocus() { showCityDropdown.value = true }
function onCityBlur() { cityBlurTimer.value = window.setTimeout(() => showCityDropdown.value = false, 200) }
function onAddressFocus() { showAddressDropdown.value = true }
function onAddressBlur() { addressBlurTimer.value = window.setTimeout(() => showAddressDropdown.value = false, 200) }

function cancelBlur(timerValue: number | null) { 
  if (timerValue) clearTimeout(timerValue) 
}

// ========== КОНЕЦ АВТОКОМПЛИТА ==========

async function loadUsers() {
  const headers = { Authorization: `Bearer ${auth.token}` }

  try {
    const [foremenRes, inspectorsRes] = await Promise.all([
      fetch(`${API_BASE}/customer/foremen-list`, { headers }),
      fetch(`${API_BASE}/customer/inspectors-list`, { headers }),
    ])

    if (foremenRes.ok) {
      foremen.value = await foremenRes.json()
    }
    if (inspectorsRes.ok) {
      inspectors.value = await inspectorsRes.json()
    }
  } catch (e) {
    console.error(e)
  }
}

async function submit() {
  submitting.value = true
  try {
    const body = {
      name: form.value.name,
      city: form.value.city,
      address: form.value.address,
      description: form.value.description,
      planned_start_date: form.value.plannedStartDate
        ? new Date(form.value.plannedStartDate).toISOString()
        : null,
      planned_end_date: form.value.plannedEndDate
        ? new Date(form.value.plannedEndDate).toISOString()
        : null,
      foreman_user_id: form.value.foremanUserId,
      inspector_user_id: form.value.inspectorUserId,
      // Используем координаты из формы, а не нули
      lat: form.value.lat,
      lng: form.value.lng,
    }

    const res = await fetch(`${API_BASE}/customer/objects`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${auth.token}`,
      },
      body: JSON.stringify(body),
    })

    if (!res.ok) {
      const data = await res.json().catch(() => ({}))
      throw new Error(data.error || 'Ошибка создания объекта')
    }

    goToObjects()
  } catch (e: any) {
    alert(e.message || 'Ошибка')
  } finally {
    submitting.value = false
  }
}

function goToObjects() {
  router.push({ name: 'customer-objects' })
}

function logout() {
  auth.clearAuth()
  router.push({ name: 'login' })
}

onMounted(loadUsers)
</script>

<style scoped>
.customer-layout {
  display: grid;
  grid-template-columns: 206px 1fr;
  min-height: 100vh;
  background: #f9fafb;
}

/* контент: симметричные отступы */
.customer-main {
  grid-column: 2;
  padding: 20px 32px;
  box-sizing: border-box;
}

.customer-header {
  margin-bottom: 16px;
}

.customer-title {
  margin: 0;
  font-size: 22px;
  font-weight: 600;
  color: #111827;
}

/* враппер, который ограничивает ширину формы и центрирует её */
.create-object-wrapper {
  display: flex;
  justify-content: center;
}

/* ширина формы */
.column {
  background: #ffffff;
  border-radius: 16px;
  padding: 16px;
  box-shadow: 0 12px 30px rgba(15, 23, 42, 0.05);
  border: 1px solid #e5e7eb;
  display: flex;
  flex-direction: column;
  width: 100%;
  max-width: 960px;
}

.column-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.column-header h2 {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
}

/* форма */
.object-form {
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.primary-btn {
padding: 8px 16px;
  background: #4f46e5;
  color: #fff;
  border: none;
  border-radius: 8px;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: background 0.2s;
  white-space: nowrap;
}
.secondary-btn {
padding: 8px 16px;
  background: #e30404;
  color: #fff;
  border: none;
  border-radius: 8px;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: background 0.2s;
  white-space: nowrap;
}

.form-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px 24px;
}

.form-field {
  display: flex;
  flex-direction: column;
  gap: 6px;
  position: relative; 
}

.form-field--full {
  grid-column: 1 / -1;
}

.form-field label {
  font-size: 13px;
  color: #6b7280;
}

.form-field input,
.form-field textarea,
.form-field select {
  border-radius: 8px;
  border: 1px solid #e5e7eb;
  padding: 8px 10px;
  font-size: 14px;
}

.form-actions {
  margin-top: 8px;
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

/* ========== Стили автокомплита ========== */
.autocomplete-list {
  position: absolute;
  top: 100%; /* Появляется сразу под полем ввода */
  left: 0;
  right: 0;
  background: #fff;
  border: 1px solid #d1d5db;
  border-top: none;
  border-radius: 0 0 8px 8px;
  max-height: 220px;
  overflow-y: auto;
  z-index: 100; /* Поверх остальных элементов */
  margin: 0;
  padding: 0;
  list-style: none;
  box-shadow: 0 8px 16px -4px rgba(0,0,0,0.1);
  animation: fadeIn 0.15s ease;
}

.autocomplete-list li {
  padding: 10px 12px;
  font-size: 14px;
  cursor: pointer;
  border-bottom: 1px solid #f3f4f6;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.autocomplete-list li:last-child {
  border-bottom: none;
}

.autocomplete-list li:hover {
  background: #f9fafb;
}

.autocomplete-status {
  color: #6b7280;
  font-style: italic;
  cursor: default;
  background: #fafafa;
}

.autocomplete-status:hover {
  background: #fafafa !important;
}

@keyframes fadeIn {
  from { opacity: 0; transform: translateY(-4px); }
  to { opacity: 1; transform: translateY(0); }
}
</style>