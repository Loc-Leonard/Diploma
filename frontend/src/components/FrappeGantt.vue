<template>
  <div ref="ganttEl" class="frappe-gantt-wrap"></div>
</template>

<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount, watch } from 'vue'
import Gantt from 'frappe-gantt'
import '@/assets/frappe-gantt.css'

type GanttTask = {
  id: string
  name: string
  start: string
  end: string
  progress: number
}

type ViewMode = 'Day' | 'Week' | 'Month' | 'Year'

const props = withDefaults(
  defineProps<{
    tasks: GanttTask[]
    viewMode?: ViewMode
    height?: number | 'auto'
  }>(),
  {
    viewMode: 'Week',
    height: 'auto',
  }
)

const ganttEl = ref<HTMLElement | null>(null)
let ganttInstance: any = null

function normalizeDate(value?: string | null): string {
  if (!value) return ''

  if (/^\d{4}-\d{2}-\d{2}$/.test(value)) {
    return value
  }

  const parsed = new Date(value)
  if (Number.isNaN(parsed.getTime())) {
    return ''
  }

  return parsed.toISOString().slice(0, 10)
}

function normalizeTasks(tasks: GanttTask[]): GanttTask[] {
  return tasks
    .map(task => {
      const start = normalizeDate(task.start)
      const end = normalizeDate(task.end)
      const progress = Math.max(0, Math.min(100, Number(task.progress ?? 0)))

      if (!start || !end) return null

      return {
        id: String(task.id),
        name: String(task.name ?? ''),
        start,
        end,
        progress,
      }
    })
    .filter((task): task is GanttTask => task !== null)
}

function buildOptions() {
  return {
    view_mode: props.viewMode,
    language: 'ru',
    date_format: 'YYYY-MM-DD',
    readonly: true,
    readonly_dates: true,
    readonly_progress: true,
    infinite_padding: false,
    scroll_to: 'start',
    container_height: props.height,
    bar_height: 24,
    padding: 18,
    popup_on: 'click',
    today_button: true,
    view_mode_select: false,
    lines: 'both',
  }
}

function clearGantt() {
  if (ganttEl.value) {
    ganttEl.value.innerHTML = ''
  }
  ganttInstance = null
}

function renderGantt() {
  if (!ganttEl.value) return

  const tasks = normalizeTasks(props.tasks)
  clearGantt()

  if (!tasks.length) return

  ganttInstance = new Gantt(ganttEl.value, tasks, buildOptions())
}

function rerenderGantt() {
  renderGantt()
}

onMounted(() => {
  renderGantt()
})

onBeforeUnmount(() => {
  clearGantt()
})

watch(
  () => props.tasks,
  () => {
    rerenderGantt()
  },
  { deep: true }
)

watch(
  () => props.viewMode,
  () => {
    rerenderGantt()
  }
)

watch(
  () => props.height,
  () => {
    rerenderGantt()
  }
)
</script>

<style>
.frappe-gantt-wrap {
  width: 100%;
  overflow-x: auto;
}

.frappe-gantt-wrap svg {
  display: block;
  min-width: 100%;
  font-family: inherit;
}
</style>