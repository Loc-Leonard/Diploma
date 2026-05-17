<template>
  <div class="modal-overlay" @click.self="$emit('close')">
    <div class="modal-card modal-card--wide">
      <div class="modal-header">
        <h2>{{ issue?.type === 'remark' ? 'Замечание' : 'Нарушение' }} #{{ issue?.id }}</h2>
        <button type="button" class="close-btn" @click="$emit('close')">×</button>
      </div>

      <div v-if="loading" class="state">Загрузка...</div>
      <div v-else-if="error" class="state state--error">{{ error }}</div>
      <template v-else-if="issue">
        <div class="detail-grid">
          <section class="detail-block">
            <div class="detail-row"><span class="label">Статус</span><IssueStatusBadge :status="issue.display_status || issue.status" /></div>
            <div class="detail-row"><span class="label">Автор</span><span>{{ issue.author_name }} ({{ issue.author_role }})</span></div>
            <div class="detail-row"><span class="label">Создано</span><span>{{ fmtDateTime(issue.created_at) }}</span></div>
            <div class="detail-row"><span class="label">Срок</span><span>{{ fmtDateTime(issue.due_date) }}</span></div>
            <div v-if="issue.classifier_code" class="detail-row"><span class="label">Классификатор</span><span>{{ issue.classifier_code }}</span></div>
            <div class="detail-row detail-row--column">
              <span class="label">Описание</span>
              <div class="description">{{ issue.description }}</div>
            </div>
          </section>

          <section class="detail-block">
            <h3>Вложения</h3>
            <div v-if="!issue.attachments?.length" class="state">Нет вложений</div>
            <div v-else class="attachments-list">
              <div v-for="file in issue.attachments" :key="file.id" class="attachment-item">
                <span>{{ file.original_file_name }}</span>
                <button type="button" class="link-btn" @click="$emit('download', file)">Скачать</button>
              </div>
            </div>

            <div class="attach-actions">
              <input ref="fileInput" type="file" class="hidden-input" @change="onFileSelect" />
              <button type="button" class="secondary-btn" @click="triggerFile">Прикрепить файл</button>
            </div>
          </section>
        </div>

        <section class="detail-block">
          <h3>Комментарии</h3>
          <div v-if="!issue.comments?.length" class="state">Комментариев пока нет</div>
          <div v-else class="timeline">
            <div v-for="comment in issue.comments" :key="comment.id" class="timeline-item">
              <div class="timeline-head">
                <strong>{{ comment.author_name }}</strong>
                <span>{{ comment.author_role }}</span>
                <span>{{ fmtDateTime(comment.created_at) }}</span>
              </div>
              <div class="timeline-body">{{ comment.comment }}</div>
            </div>
          </div>

          <div class="comment-form">
            <textarea v-model="commentText" rows="3" placeholder="Добавить комментарий"></textarea>
            <button type="button" class="primary-btn" :disabled="commentSubmitting" @click="submitComment">
              {{ commentSubmitting ? 'Отправка...' : 'Отправить комментарий' }}
            </button>
          </div>
        </section>

        <section class="detail-block">
          <h3>История статусов</h3>
          <div v-if="!issue.status_history?.length" class="state">История пуста</div>
          <div v-else class="timeline">
            <div v-for="item in issue.status_history" :key="item.id" class="timeline-item">
              <div class="timeline-head">
                <strong>{{ item.changed_by_name }}</strong>
                <span>{{ item.changed_by_role }}</span>
                <span>{{ fmtDateTime(item.created_at) }}</span>
              </div>
              <div class="timeline-body">
                {{ item.from_status || '—' }} → {{ item.to_status }}
                <template v-if="item.comment"> · {{ item.comment }}</template>
              </div>
            </div>
          </div>
        </section>

        <section v-if="canResolve" class="detail-block">
          <h3>Устранение</h3>
          <textarea v-model="resolveComment" rows="3" placeholder="Комментарий об устранении"></textarea>
          <div class="action-row">
            <button type="button" class="secondary-btn" @click="$emit('mark-in-progress', issue)">Взять в работу</button>
            <button type="button" class="primary-btn" :disabled="resolveSubmitting" @click="$emit('resolve', { issue, comment: resolveComment })">
              {{ resolveSubmitting ? 'Отправка...' : 'Отправить на проверку' }}
            </button>
          </div>
        </section>

        <section v-if="canReview" class="detail-block">
          <h3>Проверка результата</h3>
          <textarea v-model="reviewComment" rows="3" placeholder="Комментарий проверки (обязателен при отклонении)"></textarea>
          <div class="action-row">
            <button type="button" class="primary-btn" :disabled="reviewSubmitting" @click="$emit('review', { issue, decision: 'ACCEPT', comment: reviewComment })">
              Принять
            </button>
            <button type="button" class="danger-btn" :disabled="reviewSubmitting" @click="$emit('review', { issue, decision: 'REJECT', comment: reviewComment })">
              Отклонить
            </button>
          </div>
        </section>
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import IssueStatusBadge from './IssueStatusBadge.vue'

const props = defineProps<{
  issue: any | null
  loading: boolean
  error: string | null
  canResolve: boolean
  canReview: boolean
  commentSubmitting: boolean
  resolveSubmitting: boolean
  reviewSubmitting: boolean
}>()

const emit = defineEmits<{
  close: []
  download: [file: any]
  upload: [file: File]
  comment: [text: string]
  resolve: [payload: { issue: any; comment: string }]
  review: [payload: { issue: any; decision: 'ACCEPT' | 'REJECT'; comment: string }]
  'mark-in-progress': [issue: any]
}>()

const commentText = ref('')
const resolveComment = ref('')
const reviewComment = ref('')
const fileInput = ref<HTMLInputElement | null>(null)

function fmtDateTime(value?: string | null) {
  if (!value) return '—'
  return new Date(value).toLocaleString('ru-RU')
}

function triggerFile() {
  fileInput.value?.click()
}

function onFileSelect(event: Event) {
  const target = event.target as HTMLInputElement
  const file = target.files?.[0]
  if (!file) return
  emit('upload', file)
  target.value = ''
}

function submitComment() {
  const value = commentText.value.trim()
  if (!value) return
  emit('comment', value)
  commentText.value = ''
}
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(15, 23, 42, 0.55);
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 16px;
  z-index: 70;
}

.modal-card {
  width: 100%;
  background: #fff;
  border-radius: 16px;
  padding: 22px 24px 20px;
  box-shadow: 0 20px 50px rgba(15, 23, 42, 0.25);
  box-sizing: border-box;
  max-height: calc(100vh - 32px);
  overflow: auto;
}

.modal-card--wide {
  max-width: 920px;
}

.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 16px;
}

.modal-header h2 {
  margin: 0;
  font-size: 18px;
  color: #111827;
}

.close-btn {
  border: none;
  background: none;
  font-size: 22px;
  cursor: pointer;
  color: #9ca3af;
}

.detail-grid {
  display: grid;
  grid-template-columns: 1.2fr 1fr;
  gap: 16px;
}

.detail-block {
  border: 1px solid #e5e7eb;
  border-radius: 12px;
  padding: 14px;
  background: #f9fafb;
  margin-bottom: 14px;
}

.detail-block h3 {
  margin: 0 0 10px;
  font-size: 15px;
  color: #111827;
}

.detail-row {
  display: flex;
  justify-content: space-between;
  gap: 10px;
  margin-bottom: 8px;
  font-size: 13px;
  color: #374151;
}

.detail-row--column {
  flex-direction: column;
}

.label {
  color: #6b7280;
}

.description {
  color: #111827;
  white-space: pre-wrap;
}

.attachments-list,
.timeline {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.attachment-item,
.timeline-item {
  background: #fff;
  border: 1px solid #e5e7eb;
  border-radius: 10px;
  padding: 10px 12px;
}

.timeline-head {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  font-size: 12px;
  color: #6b7280;
  margin-bottom: 4px;
}

.timeline-body {
  font-size: 13px;
  color: #111827;
  white-space: pre-wrap;
}

.link-btn,
.primary-btn,
.secondary-btn,
.danger-btn {
  border: none;
  border-radius: 999px;
  cursor: pointer;
  padding: 8px 14px;
  font-size: 13px;
}

.link-btn {
  background: #eef2ff;
  color: #4338ca;
}

.primary-btn {
  background: #4f46e5;
  color: #fff;
}

.secondary-btn {
  background: #e5e7eb;
  color: #374151;
}

.danger-btn {
  background: #fee2e2;
  color: #b91c1c;
}

.comment-form,
.attach-actions {
  margin-top: 12px;
}

.comment-form textarea,
.detail-block textarea {
  width: 100%;
  box-sizing: border-box;
  border-radius: 10px;
  border: 1px solid #d1d5db;
  padding: 8px 10px;
  font-size: 14px;
  outline: none;
  background: #fff;
  resize: vertical;
  margin-bottom: 10px;
}

.action-row {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.hidden-input {
  display: none;
}

.state {
  font-size: 13px;
  color: #6b7280;
}

.state--error {
  color: #b91c1c;
}

@media (max-width: 768px) {
  .detail-grid {
    grid-template-columns: 1fr;
  }
}
</style>