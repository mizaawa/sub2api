import { computed, onBeforeUnmount, onMounted, ref } from 'vue'

export function useChartTheme() {
  const isDark = ref(
    typeof document !== 'undefined' && document.documentElement.classList.contains('dark')
  )
  let observer: MutationObserver | null = null

  const syncTheme = () => {
    isDark.value = document.documentElement.classList.contains('dark')
  }

  onMounted(() => {
    syncTheme()
    observer = new MutationObserver(syncTheme)
    observer.observe(document.documentElement, {
      attributes: true,
      attributeFilter: ['class']
    })
  })

  onBeforeUnmount(() => {
    observer?.disconnect()
  })

  return {
    distributionBorderColor: computed(() => (isDark.value ? '#82dfd8' : '#39c5bb')),
    distributionHoverBorderColor: computed(() => (isDark.value ? '#b5eee9' : '#177f79'))
  }
}
