import { createRouter, createWebHashHistory } from 'vue-router'
import { loadSession, authState } from './auth'

const NoticeEditor = () => import('./components/NoticeEditor.vue')
const page = (path, title, component, props) => ({
  path, name: path.slice(1).replaceAll('/', '-'), meta: { title }, component, props
})

const router = createRouter({
  history: createWebHashHistory(),
  routes: [
    { path: '/', redirect: '/sign-in' },
    { path: '/setup', component: () => import('./views/Setup.vue') },
    { path: '/sign-in', component: () => import('./views/Login.vue') },
    { path: '/join-family', component: () => import('./views/Register.vue') },
    {
      path: '/family-portal',
      component: () => import('./views/FamilyPortal.vue'),
      redirect: '/family-portal/overview',
      children: [
        page('/family-portal/overview', '家属首页', () => import('./views/FamilyOverview.vue')),
        page('/family-portal/vitals', '心率血压', () => import('./views/FamilyVitals.vue')),
        page('/family-portal/activity', '睡眠步数', () => import('./views/FamilyActivity.vue')),
        page('/family-portal/location', '位置信息', () => import('./views/FamilyLocation.vue')),
        page('/family-portal/profile', '个人中心', () => import('./views/FamilyProfile.vue')),
        page('/family-portal/notice', '公告内容', () => import('./views/FamilyNotice.vue'))
      ]
    },
    {
      path: '/operations',
      component: () => import('./views/OperationsShell.vue'),
      redirect: '/operations/overview',
      children: [
        page('/operations/overview', '运营总览', () => import('./views/OperationsOverview.vue')),
        page('/operations/vitals', '心率与血压', () => import('./views/VitalsOverview.vue')),
        page('/operations/activity', '睡眠与步数', () => import('./views/ActivityOverview.vue')),
        page('/operations/locations', '位置与告警', () => import('./views/LocationOverview.vue')),
        page('/operations/life', '生活信息', () => import('./views/LifeOverview.vue')),
        page('/operations/monitoring-profiles', '监测阈值档案', () => import('./views/MonitoringProfiles.vue')),
        page('/operations/residents', '老人档案', () => import('./views/Residents.vue')),
        page('/operations/guardians', '监护人信息', () => import('./views/Guardians.vue')),
        page('/operations/notices/edit', '编辑公告', NoticeEditor, { editing: true }),
        page('/operations/notices/new', '发布公告', NoticeEditor, { editing: false }),
        page('/operations/notices/detail', '公告内容', () => import('./views/NoticeDetail.vue')),
        page('/operations/notices', '公告管理', () => import('./views/NoticeManagement.vue')),
        page('/operations/families', '家属账号', () => import('./views/FamilyAccounts.vue')),
        page('/operations/comments', '评论管理', () => import('./views/CommentManagement.vue')),
        page('/operations/invitations', '家属注册邀请码', () => import('./views/FamilyInvitations.vue')),
        page('/operations/devices', '设备管理', () => import('./views/DeviceManagement.vue'))
      ]
    }
  ]
})

router.beforeEach((to, from, next) => {
  const publicPage = to.path === '/sign-in' || to.path === '/join-family'
  loadSession().then(account => {
    if (authState.setupRequired) return next(to.path === '/setup' ? undefined : '/setup')
    if (to.path === '/setup') return next('/sign-in')
    if (publicPage) return next()
    if (!account) return next('/sign-in')
    const familyPage = to.matched.some(route => route.path === '/family-portal')
    if ((account.role === 'FAMILY') !== familyPage) return next(account.role === 'FAMILY' ? '/family-portal/overview' : '/operations/overview')
    if (account.role === 'ADMIN' && ['caregiver', 'auditor'].includes(account.duty) && !['/operations/overview', '/operations'].includes(to.path)) return next('/operations/overview')
    next()
  }).catch(error => {
    if (error.response?.data?.code === 'DATABASE_UNAVAILABLE' && authState.account) {
      const familyPage = to.matched.some(route => route.path === '/family-portal')
      if (!publicPage && to.path !== '/setup' && (authState.account.role === 'FAMILY') !== familyPage) return next(authState.account.role === 'FAMILY' ? '/family-portal/overview' : '/operations/overview')
      return next(to.path === '/setup' ? '/sign-in' : undefined)
    }
    next(publicPage || (to.path === '/setup' && error.response?.data?.code === 'DATABASE_UNAVAILABLE') ? undefined : '/sign-in')
  })
})

export default router
