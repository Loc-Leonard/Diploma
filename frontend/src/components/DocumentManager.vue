<template>
  <div class="document-manager">
    <!-- Заголовок с кнопкой загрузки -->
    <div class="document-header">
      <h3 class="document-title">📄 Документы</h3>
      <div class="document-actions">
        <button
          type="button"
          class="upload-btn"
          @click="$emit('upload-click')"
          :disabled="uploading"
        >
          {{ uploading ? 'Загрузка...' : '+ Добавить документ' }}
        </button>
        <input
          type="file"
          ref="fileInput"
          @change="$emit('file-select', $event)"
          accept=".jpg,.jpeg,.png,.gif,.pdf,.doc,.docx,.xls,.xlsx"
          class="file-input-hidden"
        />
      </div>
    </div>

    <!-- Прогресс загрузки -->
    <div v-if="uploading" class="upload-progress">
      <div class="progress-bar">
        <div class="progress-fill"></div>
      </div>
      <span class="progress-text">Загрузка и обработка CV...</span>
    </div>

    <!-- Список документов -->
    <div v-if="documents.length === 0 && !loading && !error" class="empty-state">
      <p>📭 Нет документов</p>
      <p class="empty-hint">Загрузите первый документ</p>
    </div>

    <div v-else-if="loading" class="loading-state">
      <div class="loading-spinner"></div>
      <span>Загрузка документов...</span>
    </div>

    <div v-else-if="error" class="error-state">
      <p>⚠️ {{ error }}</p>
      <button @click="$emit('retry')" class="retry-btn">Попробовать снова</button>
    </div>

    <div v-else class="documents-list">
      <div
        v-for="doc in documents"
        :key="doc.id"
        class="document-item"
        :class="`document-type--${doc.document_type.toLowerCase()}`"
      >
        <div class="doc-icon">
          <span v-if="isImage(doc.mime_type)">🖼️</span>
          <span v-else-if="isPDF(doc.mime_type)">📕</span>
          <span v-else-if="isWord(doc.mime_type)">📘</span>
          <span v-else-if="isExcel(doc.mime_type)">📗</span>
          <span v-else>📄</span>
        </div>

        <div class="doc-info">
          <div class="doc-name">{{ doc.original_file_name }}</div>
          <div class="doc-meta">
            <span class="doc-type">{{ getDocumentTypeLabel(doc.document_type) }}</span>
            <span class="doc-date">{{ fmtDate(doc.created_at) }}</span>
            <span class="doc-uploader">👤 {{ doc.uploaded_by }}</span>
          </div>
          <div class="doc-cv-status" :class="`cv-status--${doc.cv_status.toLowerCase()}`">
            <span class="cv-indicator"></span>
            <span class="cv-text">{{ getCVStatusLabel(doc.cv_status) }}</span>
            <span v-if="doc.cv_confidence > 0" class="cv-confidence">
              ({{ Math.round(doc.cv_confidence * 100) }}%)
            </span>
          </div>
        </div>

        <div class="doc-actions">
          <button
            @click="$emit('download', doc)"
            class="action-btn download"
            title="Скачать"
            :disabled="downloadingId === doc.id"
          >
            {{ downloadingId === doc.id ? '⏳' : '⬇️' }}
          </button>
          <button
            v-if="canDelete(doc)"
            @click="$emit('delete', doc)"
            class="action-btn delete"
            title="Удалить"
            :disabled="deletingId === doc.id"
          >
            {{ deletingId === doc.id ? '⏳' : '🗑️' }}
          </button>
        </div>
      </div>
    </div>

    <!-- Сообщение об ошибке -->
    <div v-if="errorMessage" class="error-message">
      <span>❌ {{ errorMessage }}</span>
      <button @click="$emit('clear-error')" class="close-error">✕</button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useAuthStore } from '../stores/auth'

interface Document {
  id: number
  document_type: string
  original_file_name: string
  mime_type: string
  cv_status: string
  cv_confidence: number
  created_at: string
  uploaded_by: string
}

interface Props {
  documents: Document[]
  loading: boolean
  uploading: boolean
  error: string | null
  errorMessage: string | null
  deletingId: number | null
  downloadingId: number | null
  canDeleteDoc: (doc: Document) => boolean
}

const props = defineProps<Props>()

defineEmits<{
  'upload-click': []
  'file-select': [event: Event]
  'retry': []
  'download': [doc: Document]
  'delete': [doc: Document]
  'clear-error': []
}>()

const auth = useAuthStore()
const fileInput = ref<HTMLInputElement | null>(null)

// Проверка типа файла
function isImage(mimeType: string) {
  return mimeType?.startsWith('image/')
}

function isPDF(mimeType: string) {
  return mimeType === 'application/pdf'
}

function isWord(mimeType: string) {
  return mimeType?.includes('word') || mimeType?.includes('document')
}

function isExcel(mimeType: string) {
  return mimeType?.includes('excel') || mimeType?.includes('spreadsheet')
}

// Форматирование даты
function fmtDate(dateString: string) {
  if (!dateString) return '—'
  return new Date(dateString).toLocaleDateString('ru-RU', {
    day: '2-digit',
    month: '2-digit',
    year: '2-digit'
  })
}

// Labels
function getDocumentTypeLabel(type: string) {
  const labels: Record<string, string> = {
    'TTN': '🚚 ТТН',
    'QUALITY_PASSPORT': '📋 Паспорт качества',
    'PHOTO': '📷 Фото',
    'OTHER': '📄 Другое'
  }
  return labels[type] ?? type
}

function getCVStatusLabel(status: string) {
  const labels: Record<string, string> = {
    'PENDING': 'Обработка...',
    'DONE': 'Обработан',
    'FAILED': 'Ошибка CV'
  }
  return labels[status] ?? status
}

function canDelete(doc: Document) {
  return props.canDeleteDoc(doc)
}
</script>

<style scoped>
.document-manager {
  background: #fff;
  border-radius: 12px;
  padding: 16px;
  border: 1px solid #e5e7eb;
}

.document-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.document-title {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: #111827;
}

.document-actions {
  display: flex;
  align-items: center;
  gap: 8px;
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
}

.upload-btn:hover:not(:disabled) {
  background: #4338ca;
}

.upload-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.file-input-hidden {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border: 0;
}

.upload-progress {
  margin-bottom: 16px;
  padding: 12px;
  background: #eff6ff;
  border-radius: 8px;
  border: 1px solid #bfdbfe;
}

.progress-bar {
  height: 4px;
  background: #dbeafe;
  border-radius: 2px;
  overflow: hidden;
  margin-bottom: 8px;
}

.progress-fill {
  height: 100%;
  width: 60%;
  background: #3b82f6;
  animation: pulse 1.5s infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

.progress-text {
  font-size: 12px;
  color: #1e40af;
}

.empty-state {
  text-align: center;
  padding: 32px 16px;
  color: #6b7280;
}

.empty-hint {
  font-size: 12px;
  color: #9ca3af;
  margin-top: 4px;
}

.loading-state {
  text-align: center;
  padding: 24px;
  color: #6b7280;
  font-size: 14px;
}

.error-state {
  text-align: center;
  padding: 24px;
  color: #b91c1c;
}

.retry-btn {
  margin-top: 8px;
  padding: 6px 12px;
  background: #fef2f2;
  border: 1px solid #fecaca;
  border-radius: 6px;
  color: #b91c1c;
  font-size: 13px;
  cursor: pointer;
}

.retry-btn:hover {
  background: #fee2e2;
}

.documents-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.document-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  background: #f9fafb;
  border-radius: 8px;
  border: 1px solid #e5e7eb;
  transition: border-color 0.2s;
}

.document-item:hover {
  border-color: #4f46e5;
}

.doc-icon {
  font-size: 24px;
  flex-shrink: 0;
}

.doc-info {
  flex: 1;
  min-width: 0;
}

.doc-name {
  font-size: 14px;
  font-weight: 500;
  color: #111827;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  margin-bottom: 4px;
}

.doc-meta {
  display: flex;
  gap: 8px;
  font-size: 11px;
  color: #6b7280;
  margin-bottom: 4px;
}

.doc-type {
  background: #e5e7eb;
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 10px;
}

.doc-cv-status {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 11px;
}

.cv-indicator {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #9ca3af;
}

.cv-status--pending .cv-indicator {
  background: #f59e0b;
  animation: blink 1s infinite;
}

.cv-status--done .cv-indicator {
  background: #10b981;
}

.cv-status--failed .cv-indicator {
  background: #ef4444;
}

@keyframes blink {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.3; }
}

.cv-text {
  color: #6b7280;
}

.cv-confidence {
  color: #10b981;
  font-weight: 500;
}

.doc-actions {
  display: flex;
  gap: 8px;
  flex-shrink: 0;
}

.action-btn {
  width: 32px;
  height: 32px;
  border-radius: 6px;
  border: 1px solid #e5e7eb;
  background: #fff;
  cursor: pointer;
  font-size: 14px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
}

.action-btn:hover {
  background: #f3f4f6;
  border-color: #d1d5db;
}

.action-btn.delete:hover {
  background: #fef2f2;
  border-color: #ef4444;
  color: #ef4444;
}

.action-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.error-message {
  margin-top: 16px;
  padding: 12px;
  background: #fef2f2;
  border: 1px solid #fecaca;
  border-radius: 8px;
  color: #b91c1c;
  font-size: 13px;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.close-error {
  background: none;
  border: none;
  cursor: pointer;
  font-size: 16px;
  color: #b91c1c;
  padding: 0;
  line-height: 1;
}

.close-error:hover {
  color: #991b1b;
}

/* Адаптивность */
@media (max-width: 768px) {
  .document-header {
    flex-direction: column;
    gap: 12px;
    align-items: stretch;
  }

  .upload-btn {
    text-align: center;
  }

  .document-item {
    flex-wrap: wrap;
  }

  .doc-info {
    min-width: calc(100% - 60px);
  }

  .doc-actions {
    width: 100%;
    justify-content: flex-end;
    margin-top: 8px;
  }
}
</style>