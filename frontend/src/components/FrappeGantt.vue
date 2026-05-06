<template>
    <div ref="ganttEl" class="frappe-gantt-wrap"></div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import Gantt from 'frappe-gantt'


const props = defineProps<{
    tasks: {
        id: string
        name: string
        start: string
        end: string
        progress: number
    }[]
    veiwMode?: string
}>()

const ganttEl = ref<HTMLElement | null>(null)
let ganttInstance: any = null

onMounted(() => {
    if (!ganttEl.value || !props.tasks.length) return

    ganttInstance = new Gantt(ganttEl.value, props.tasks, {
        view_mode: props.veiwMode ?? 'Week',
        language: 'ru',
        date_format: 'YYYY-MM-DD',
    })
})

watch(() => props.tasks, (NewTasks) => {
    if (!ganttInstance || !NewTasks.length) return
    ganttInstance.refresh(NewTasks)
}, { deep: true })
</script>

<style>
.frappe-gantt-wrap {
    overflow-x: auto;
    width: 100%;
}

.frappe-gantt-wrap svg {
    font-family: inherit;
}
</style>
