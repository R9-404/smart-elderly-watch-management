import { onBeforeUnmount, onMounted, ref } from 'vue'

// Keep this breakpoint aligned with the collection layouts in commercial.css.
export function useCollectionLayout () {
  const query = window.matchMedia('(max-width: 1100px)')
  const mobileLayout = ref(query.matches)
  const update = event => { mobileLayout.value = event.matches }
  onMounted(() => {
    query.addEventListener('change', update)
    mobileLayout.value = query.matches
  })
  onBeforeUnmount(() => query.removeEventListener('change', update))
  return mobileLayout
}
