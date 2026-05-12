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

function normalizeTasks(tasks: GanttTask[]): GanttTask[] {
  return tasks
    .filter(task => task.start && task.end)
    .map(task => ({
      ...task,
      start: task.start.slice(0, 10),
      end: task.end.slice(0, 10),
      progress: Math.max(0, Math.min(100, Number(task.progress ?? 0))),
    }))
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

function renderGantt() {
  if (!ganttEl.value) return

  const tasks = normalizeTasks(props.tasks)
  ganttEl.value.innerHTML = ''

  if (!tasks.length) {
    ganttInstance = null
    return
  }

  ganttInstance = new Gantt(ganttEl.value, tasks, buildOptions())
}

function refreshTasks() {
  if (!ganttEl.value) return

  const tasks = normalizeTasks(props.tasks)

  if (!tasks.length) {
    ganttEl.value.innerHTML = ''
    ganttInstance = null
    return
  }

  if (!ganttInstance) {
    renderGantt()
    return
  }

  ganttInstance.refresh(tasks)
  ganttInstance.update_options(buildOptions())
  ganttInstance.change_view_mode(props.viewMode, true)
}

onMounted(() => {
  renderGantt()
})

onBeforeUnmount(() => {
  if (ganttEl.value) {
    ganttEl.value.innerHTML = ''
  }
  ganttInstance = null
})

watch(
  () => props.tasks,
  () => {
    refreshTasks()
  },
  { deep: true }
)

watch(
  () => props.viewMode,
  newMode => {
    if (!ganttInstance) {
      renderGantt()
      return
    }
    ganttInstance.change_view_mode(newMode, true)
  }
)

watch(
  () => props.height,
  () => {
    if (!ganttInstance) {
      renderGantt()
      return
    }
    ganttInstance.update_options(buildOptions())
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