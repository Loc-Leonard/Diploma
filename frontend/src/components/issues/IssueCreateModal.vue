<template>
  <div class="modal-overlay" @click.self="$emit('close')">
    <div class="modal-card">
      <h2>{{ mode === 'remark' ? 'Новое замечание' : 'Новое нарушение' }}</h2>

      <div class="form-field">
        <label>Заголовок</label>
        <input v-model="form.title" type="text" placeholder="Короткий заголовок" />
      </div>

      <div v-if="mode === 'violation'" class="form-field">
        <label>Классификатор нарушения <span class="required">*</span></label>
        <input
          v-model="form.classifier_code"
          type="text"
          placeholder="Например: KO-12.4"
        />
      </div>

      <div class="form-field">
        <label>Описание <span class="required">*</span></label>
        <textarea v-model="form.description" rows="4" />
      </div>

      <div class="form-field">
        <label>Срок устранения <span class="required">*</span></label>
        <input v-model="form.due_date" type="datetime-local" />
      </div>

      <div class="form-field">
        <label>Комментарий</label>
        <textarea v-model="form.comment" rows="3" />
      </div>

      <div class="form-field">
        <label>Вложения</label>
        <input
          ref="fileInputRef"
          type="file"
          multiple
          accept=".jpg,.jpeg,.png,.gif,.pdf,.doc,.docx,.xls,.xlsx"
          @change="handleFileChange"
        />
        <div v-if="selectedFiles.length" class="files-list">
          <div v-for="file in selectedFiles" :key="fileKey(file)" class="file-chip">
            <span class="file-name">{{ file.name }}</span>
            <button type="button" class="remove-file-btn" @click="removeFile(file)">
              ×
            </button>
          </div>
        </div>
      </div>

      <div v-if="localError" class="state state--error">{{ localError }}</div>
      <div v-else-if="error" class="state state--error">{{ error }}</div>

      <div class="modal-actions">
        <button type="button" class="secondary-btn" @click="$emit('close')">Отмена</button>
        <button type="button" class="primary-btn" :disabled="submitting" @click="submit">
          {{ submitting ? 'Сохранение...' : 'Создать' }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref, watch } from 'vue'

const props = defineProps<{
  mode: 'remark' | 'violation'
  submitting: boolean
  error: string | null
}>()

const emit = defineEmits<{
  close: []
  submit: [payload: {
    title: string
    description: string
    due_date: string | null
    classifier_code: string | null
    comment: string
    files: File[]
  }]
}>()

const fileInputRef = ref<HTMLInputElement | null>(null)
const selectedFiles = ref<File[]>([])
const localError = ref<string | null>(null)

const form = reactive({
  title: '',
  description: '',
  due_date: '',
  classifier_code: '',
  comment: '',
})

watch(
  () => props.mode,
  () => resetForm(),
  { immediate: true },
)

function resetForm() {
  form.title = ''
  form.description = ''
  form.due_date = ''
  form.classifier_code = ''
  form.comment = ''
  selectedFiles.value = []
  localError.value = null
  if (fileInputRef.value) {
    fileInputRef.value.value = ''
  }
}

function handleFileChange(event: Event) {
  const target = event.target as HTMLInputElement
  const files = Array.from(target.files ?? [])

  const allowedTypes = [
    'image/jpeg',
    'image/png',
    'image/jpg',
    'image/gif',
    'application/pdf',
    'application/msword',
    'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
    'application/vnd.ms-excel',
    'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
  ]

  const maxSize = 10 * 1024 * 1024

  for (const file of files) {
    if (!allowedTypes.includes(file.type)) {
      localError.value = `Файл "${file.name}" имеет неподдерживаемый формат`
      continue
    }
    if (file.size > maxSize) {
      localError.value = `Файл "${file.name}" превышает 10MB`
      continue
    }

    const exists = selectedFiles.value.some(
      (f) => f.name === file.name && f.size === file.size && f.lastModified === file.lastModified,
    )
    if (!exists) {
      selectedFiles.value.push(file)
    }
  }

  if (target) {
    target.value = ''
  }
}

function removeFile(file: File) {
  selectedFiles.value = selectedFiles.value.filter(
    (f) => !(f.name === file.name && f.size === file.size && f.lastModified === file.lastModified),
  )
}

function fileKey(file: File) {
  return `${file.name}-${file.size}-${file.lastModified}`
}

function submit() {
  localError.value = null

  if (!form.description.trim()) {
    localError.value = 'Заполните описание'
    return
  }

  if (!form.due_date) {
    localError.value = 'Укажите срок устранения'
    return
  }

  if (props.mode === 'violation' && !form.classifier_code.trim()) {
    localError.value = 'Укажите классификатор нарушения'
    return
  }

  emit('submit', {
    title: form.title.trim(),
    description: form.description.trim(),
    due_date: form.due_date ? new Date(form.due_date).toISOString() : null,
    classifier_code: props.mode === 'violation' ? form.classifier_code.trim() || null : null,
    comment: form.comment.trim(),
    files: selectedFiles.value,
  })
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
  z-index: 60;
}

.modal-card {
  width: 100%;
  max-width: 560px;
  background: #fff;
  border-radius: 16px;
  padding: 22px 24px 20px;
  box-shadow: 0 20px 50px rgba(15, 23, 42, 0.25);
  box-sizing: border-box;
  max-height: calc(100vh - 32px);
  overflow-y: auto;
}

.modal-card h2 {
  margin: 0 0 16px;
  font-size: 18px;
  font-weight: 600;
  color: #111827;
}

.form-field {
  display: flex;
  flex-direction: column;
  gap: 6px;
  margin-bottom: 12px;
}

.form-field label {
  font-size: 13px;
  color: #6b7280;
}

.form-field input,
.form-field textarea {
  border-radius: 10px;
  border: 1px solid #d1d5db;
  padding: 8px 10px;
  font-size: 14px;
  outline: none;
  background: #f9fafb;
}

.form-field textarea {
  resize: vertical;
  min-height: 84px;
}

.required {
  color: #dc2626;
}

.files-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 4px;
}

.file-chip {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  background: #eef2ff;
  color: #3730a3;
  border-radius: 999px;
  padding: 6px 10px;
  font-size: 12px;
  max-width: 100%;
}

.file-name {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.remove-file-btn {
  border: none;
  background: transparent;
  color: #4338ca;
  cursor: pointer;
  font-size: 14px;
  line-height: 1;
  padding: 0;
}

.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  margin-top: 4px;
}

.primary-btn,
.secondary-btn {
  padding: 8px 16px;
  border-radius: 999px;
  border: none;
  font-size: 14px;
  cursor: pointer;
}

.primary-btn {
  background: #4f46e5;
  color: #fff;
}

.primary-btn:disabled {
  opacity: 0.5;
  cursor: default;
}

.secondary-btn {
  background: #e5e7eb;
  color: #374151;
}

.state {
  font-size: 13px;
  color: #6b7280;
}

.state--error {
  color: #b91c1c;
}
</style>