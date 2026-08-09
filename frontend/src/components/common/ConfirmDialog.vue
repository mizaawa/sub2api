<template>
  <BaseDialog :show="show" :title="title" width="narrow" @close="handleCancel">
    <div class="space-y-4">
      <p class="confirm-dialog-message">{{ message }}</p>
      <slot></slot>
    </div>

    <template #footer>
      <div class="flex justify-end gap-3">
        <button
          @click="handleCancel"
          type="button"
          class="btn btn-secondary"
          :disabled="loading"
        >
          {{ cancelText }}
        </button>
        <button
          @click="handleConfirm"
          type="button"
          :class="['btn', danger ? 'btn-danger' : 'btn-primary']"
          :disabled="loading"
        >
          <span v-if="loading" class="confirm-dialog-spinner" aria-hidden="true"></span>
          {{ confirmText }}
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseDialog from './BaseDialog.vue'

const { t } = useI18n()

interface Props {
  show: boolean
  title: string
  message: string
  confirmText?: string
  cancelText?: string
  danger?: boolean
  loading?: boolean
}

interface Emits {
  (e: 'confirm'): void
  (e: 'cancel'): void
}

const props = withDefaults(defineProps<Props>(), {
  danger: false,
  loading: false
})

const confirmText = computed(() => props.confirmText || t('common.confirm'))
const cancelText = computed(() => props.cancelText || t('common.cancel'))

const emit = defineEmits<Emits>()

const handleConfirm = () => {
  if (props.loading) return
  emit('confirm')
}

const handleCancel = () => {
  if (props.loading) return
  emit('cancel')
}
</script>

<style scoped>
.confirm-dialog-message { color: var(--md-sys-color-on-surface-variant); }
.confirm-dialog-spinner {
  width: 0.9rem;
  height: 0.9rem;
  border: 2px solid currentColor;
  border-right-color: transparent;
  border-radius: 999px;
  animation: confirm-dialog-spin 0.7s linear infinite;
}
@keyframes confirm-dialog-spin { to { transform: rotate(360deg); } }
</style>
