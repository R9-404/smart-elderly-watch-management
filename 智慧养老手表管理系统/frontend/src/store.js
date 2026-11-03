import { reactive } from 'vue'
const state = reactive({ openTab: [], activeIndex: '/operations/overview' })
export default {
  state,
  addTab (tab) { state.openTab.push(tab) },
  removeTab (route) { state.openTab = state.openTab.filter(tab => tab.route !== route) },
  clearTabs () { state.openTab = []; state.activeIndex = '/operations/overview' },
  activate (route) { state.activeIndex = route }
}
