<template>
  <section class="issues-section">
    <div class="section-header">
      <h2>Замечания и нарушения</h2>

      <div class="section-actions">
        <div class="issue-filters">
          <button
            v-for="item in filters"
            :key="item.value"
            type="button"
            class="filter-pill"
            :class="{ 'filter-pill--active': activeFilter === item.value }"
            @click="$emit('change-filter', item.value)"
          >
            {{ item.label }}
          </button>
        </div>

        <button
          v-if="canCreate"
          type="button"
          class="upload-btn"
          @click="$emit('create')"
        >
          {{ createButtonLabel }}
        </button>
      </div>
    </div>

    <div v-if="loading" class="state">Загрузка замечаний и нарушений...</div>
    <div v-else-if="error" class="state state--error">{{ error }}</div>
    <div v-else-if="issues.length === 0" class="state">Записей пока нет.</div>

    <div v-else class="issues-list">
      <article
        v-for="issue in issues"
        :key="issue.id"
        class="issue-card"
        @click="$emit('open', issue)"
      >
        <div class="issue-card-top">
          <div class="issue-card-main">
            <div class="issue-card-title-row">
              <span class="issue-type" :class="`issue-type--${issue.type}`">
                {{ issue.type === 'remark' ? 'Замечание' : 'Нарушение' }}
              </span>
              <IssueStatusBadge :status="issue.display_status || issue.status" />
            </div>

            <div class="issue-title">
              {{ issue.title || issue.description }}
            </div>

            <div class="issue-description">
              {{ issue.description }}
            </div>
          </div>
        </div>

        <div class="issue-meta">
          <span>Автор: {{ issue.author_name }}</span>
          <span>Роль: {{ issue.author_role }}</span>
          <span>Создано: {{ fmtDateTime(issue.created_at) }}</span>
          <span v-if="issue.due_date">Срок: {{ fmtDate(issue.due_date) }}</span>
          <span v-if="issue.attachments?.length">Вложений: {{ issue.attachments.length }}</span>
          <span v-if="issue.classifier_code">Классификатор: {{ issue.classifier_code }}</span>
        </div>
      </article>
    </div>
  </section>
</template>

<script setup lang="ts">
import IssueStatusBadge from './IssueStatusBadge.vue'

defineProps<{
  issues: any[]
  loading: boolean
  error: string | null
  canCreate: boolean
  createButtonLabel: string
  activeFilter: string
}>()

defineEmits<{
  create: []
  open: [issue: any]
  'change-filter': [value: string]
}>()

const filters = [
  { value: 'all', label: 'Все' },
  { value: 'remark', label: 'Замечания' },
  { value: 'violation', label: 'Нарушения' },
  { value: 'open', label: 'Открытые' },
  { value: 'overdue', label: 'Просроченные' },
  { value: 'resolved_by_foreman', label: 'На проверке' },
  { value: 'closed', label: 'Закрытые' },
]

function fmtDate(value?: string | null) {
  if (!value) return '—'
  return new Date(value).toLocaleDateString('ru-RU')
}

function fmtDateTime(value?: string | null) {
  if (!value) return '—'
  return new Date(value).toLocaleString('ru-RU')
}
</script>

<style scoped>
.section-actions{
  display: flex;
  gap: 10px;
  align-items: flex-start;
  justify-content: flex-end;
  flex: 1;
  flex-wrap: wrap;
}
.issues-section {
  background: #fff;
  border: 1px solid #e5e7eb;
  border-radius: 16px;
  padding: 16px 20px;
  margin-bottom: 16px;
  box-shadow: 0 12px 30px rgba(15, 23, 42, 0.05);
}

.section-header {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  align-items: flex-start;
  margin-bottom: 14px;
  flex-wrap: wrap;
}

.section-header h2 {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: #111827;
}

.upload-btn {
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
  flex-shrink: 0;
}

.upload-btn:hover:not(:disabled) {
  background: #4338ca;
}

.upload-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.issue-filters {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  justify-content: flex-end;
  flex: 1;
}

.filter-pill {
  border: 1px solid #d1d5db;
  background: #fff;
  color: #374151;
  border-radius: 999px;
  padding: 6px 12px;
  font-size: 12px;
  cursor: pointer;
}

.filter-pill--active {
  background: #eef2ff;
  border-color: #a5b4fc;
  color: #4338ca;
}

.issues-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.issue-card {
  border: 1px solid #e5e7eb;
  border-radius: 12px;
  padding: 14px;
  background: #f9fafb;
  cursor: pointer;
  transition: box-shadow 0.15s, border-color 0.15s;
}

.issue-card:hover {
  border-color: #a5b4fc;
  box-shadow: 0 6px 18px rgba(15, 23, 42, 0.06);
}

.issue-card-title-row {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  align-items: center;
  margin-bottom: 8px;
}

.issue-type {
  display: inline-flex;
  align-items: center;
  padding: 3px 10px;
  border-radius: 999px;
  font-size: 11px;
  font-weight: 600;
}

.issue-type--remark {
  background: #e0f2fe;
  color: #0369a1;
}

.issue-type--violation {
  background: #fef2f2;
  color: #b91c1c;
}

.issue-title {
  font-size: 14px;
  font-weight: 600;
  color: #111827;
}

.issue-description {
  margin-top: 6px;
  font-size: 13px;
  color: #4b5563;
}

.issue-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 8px 14px;
  margin-top: 12px;
  font-size: 12px;
  color: #6b7280;
}

.state {
  font-size: 13px;
  color: #6b7280;
}

.state--error {
  color: #b91c1c;
}

@media (max-width: 768px) {
  .section-actions {
    align-items: stretch;
    width: 100%;
  }

  .issue-filters {
    justify-content: flex-start;
  }
}
</style>