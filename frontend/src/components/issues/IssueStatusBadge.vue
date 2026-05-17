<template>
  <span class="issue-status-chip" :class="statusClass">
    {{ statusLabel }}
  </span>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  status: string
}>()

const statusLabel = computed(() => {
  const map: Record<string, string> = {
    open: 'Открыто',
    in_progress: 'В работе',
    resolved_by_foreman: 'На проверке',
    accepted: 'Принято',
    rejected: 'Отклонено',
    overdue: 'Просрочено',
  }
  return map[props.status] ?? props.status
})

const statusClass = computed(() => {
  return {
    'issue-status-chip--open': props.status === 'open',
    'issue-status-chip--in-progress': props.status === 'in_progress',
    'issue-status-chip--resolved': props.status === 'resolved_by_foreman',
    'issue-status-chip--accepted': props.status === 'accepted',
    'issue-status-chip--rejected': props.status === 'rejected',
    'issue-status-chip--overdue': props.status === 'overdue',
  }
})
</script>

<style scoped>
.issue-status-chip {
  display: inline-flex;
  align-items: center;
  padding: 3px 10px;
  border-radius: 999px;
  font-size: 11px;
  font-weight: 600;
  white-space: nowrap;
}

.issue-status-chip--open {
  background: #dbeafe;
  color: #1d4ed8;
}

.issue-status-chip--in-progress {
  background: #fef3c7;
  color: #92400e;
}

.issue-status-chip--resolved {
  background: #ede9fe;
  color: #6d28d9;
}

.issue-status-chip--accepted {
  background: #dcfce7;
  color: #166534;
}

.issue-status-chip--rejected {
  background: #fee2e2;
  color: #b91c1c;
}

.issue-status-chip--overdue {
  background: #fecaca;
  color: #991b1b;
}
</style>