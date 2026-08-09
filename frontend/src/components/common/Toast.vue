<template>
  <Teleport to="body">
    <div
      class="pointer-events-none fixed right-4 top-4 z-[9999] space-y-3"
      aria-live="polite"
      aria-atomic="true"
    >
      <TransitionGroup
        enter-active-class="transition ease-out duration-300"
        enter-from-class="opacity-0 translate-x-full"
        enter-to-class="opacity-100 translate-x-0"
        leave-active-class="transition ease-in duration-200"
        leave-from-class="opacity-100 translate-x-0"
        leave-to-class="opacity-0 translate-x-full"
      >
        <div
          v-for="toast in toasts"
          :key="toast.id"
          :class="[
            'toast-card pointer-events-auto w-[min(24rem,calc(100vw-2rem))] overflow-hidden rounded-2xl',
            getToastTone(toast.type)
          ]"
        >
          <div class="p-4">
            <div class="flex items-start gap-3">
              <!-- Icon -->
              <div class="mt-0.5 flex-shrink-0">
                <Icon
                  :name="getToastIconName(toast.type)"
                  size="md"
                  :class="getIconColor(toast.type)"
                  aria-hidden="true"
                />
              </div>

              <!-- Content -->
              <div class="min-w-0 flex-1">
                <p v-if="toast.title" class="toast-title">
                  {{ toast.title }}
                </p>
                <p
                  :class="[
                    'text-sm leading-relaxed',
                    toast.title
                      ? 'mt-1 toast-message-muted'
                      : 'toast-message'
                  ]"
                >
                  {{ toast.message }}
                </p>
              </div>

              <!-- Close button -->
              <button
                @click="removeToast(toast.id)"
                class="toast-close -m-1 flex-shrink-0 rounded-lg p-1 transition-colors"
                aria-label="Close notification"
              >
                <Icon name="x" size="sm" />
              </button>
            </div>
          </div>

          <!-- Progress bar -->
          <div v-if="toast.duration" class="toast-progress-track h-1">
            <div
              :class="['h-full toast-progress', getProgressBarColor(toast.type)]"
              :style="{ animationDuration: `${toast.duration}ms` }"
            ></div>
          </div>
        </div>
      </TransitionGroup>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import Icon from '@/components/icons/Icon.vue'
import { useAppStore } from '@/stores/app'

const appStore = useAppStore()

const toasts = computed(() => appStore.toasts)

const getToastIconName = (type: string): 'checkCircle' | 'xCircle' | 'exclamationTriangle' | 'infoCircle' => {
  switch (type) {
    case 'success':
      return 'checkCircle'
    case 'error':
      return 'xCircle'
    case 'warning':
      return 'exclamationTriangle'
    case 'info':
    default:
      return 'infoCircle'
  }
}

const getIconColor = (type: string): string => {
  const colors: Record<string, string> = {
    success: 'text-green-500',
    error: 'text-red-500',
    warning: 'text-yellow-500',
    info: 'text-blue-500'
  }
  return colors[type] || colors.info
}

const getToastTone = (type: string): string => `toast-tone-${type || 'info'}`

const getProgressBarColor = (type: string): string => {
  const colors: Record<string, string> = {
    success: 'bg-green-500',
    error: 'bg-red-500',
    warning: 'bg-yellow-500',
    info: 'bg-blue-500'
  }
  return colors[type] || colors.info
}

const removeToast = (id: string) => {
  appStore.hideToast(id)
}
</script>

<style scoped>
.toast-card {
  border: 1px solid var(--md-sys-color-outline);
  background: var(--md-sys-color-surface-container-high);
  color: var(--md-sys-color-on-surface);
  box-shadow: 0 1rem 2rem rgb(6 54 51 / 0.16);
}
.toast-title, .toast-message { color: var(--md-sys-color-on-surface); }
.toast-message-muted { color: var(--md-sys-color-on-surface-variant); }
.toast-close { color: var(--md-sys-color-on-surface-variant); }
.toast-close:hover { background: color-mix(in srgb, var(--md-sys-color-primary) 12%, transparent); color: var(--md-sys-color-on-surface); }
.toast-progress-track { background: color-mix(in srgb, var(--md-sys-color-outline) 35%, transparent); }
.toast-tone-success { border-left: 4px solid #39c5bb; }
.toast-tone-error { border-left: 4px solid #ef6f7b; }
.toast-tone-warning { border-left: 4px solid #e9c46a; }
.toast-tone-info { border-left: 4px solid var(--md-sys-color-primary); }
.toast-progress.bg-green-500 { background: #39c5bb; }
.toast-progress.bg-blue-500 { background: var(--md-sys-color-primary); }
.toast-progress {
  width: 100%;
  animation-name: toast-progress-shrink;
  animation-timing-function: linear;
  animation-fill-mode: forwards;
}

@keyframes toast-progress-shrink {
  from {
    width: 100%;
  }
  to {
    width: 0%;
  }
}
</style>
