import { reactive } from 'vue'

export interface Tab { route: string; name: string }

const state = reactive<{ openTab: Tab[]; activeIndex: string }>({ openTab: [], activeIndex: '/operations/overview' })

export default {
  state,
  addTab (tab: Tab) { state.openTab.push(tab) },
  removeTab (route: string) { state.openTab = state.openTab.filter(tab => tab.route !== route) },
  clearTabs () { state.openTab = []; state.activeIndex = '/operations/overview' },
  activate (route: string) { state.activeIndex = route }
}
