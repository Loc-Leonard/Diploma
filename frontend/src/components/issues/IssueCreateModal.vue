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

      <div v-if="error" class="state state--error">{{ error }}</div>

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
import { reactive, watch } from 'vue'

const props = defineProps<{
  mode: 'remark' | 'violation'
  submitting: boolean
  error: string | null
}>()

const emit = defineEmits<{
  close: []
  submit: [payload: any]
}>()

const form = reactive({
  title: '',
  description: '',
  due_date: '',
  classifier_code: '',
  comment: '',
})

watch(
  () => props.mode,
  () => {
    form.title = ''
    form.description = ''
    form.due_date = ''
    form.classifier_code = ''
    form.comment = ''
  },
  { immediate: true },
)

function submit() {
  emit('submit', {
    title: form.title.trim(),
    description: form.description.trim(),
    due_date: form.due_date ? new Date(form.due_date).toISOString() : null,
    classifier_code: form.classifier_code.trim() || null,
    comment: form.comment.trim(),
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