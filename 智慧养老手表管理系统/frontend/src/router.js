import { createRouter, createWebHashHistory } from 'vue-router'
import { loadSession, authState } from './auth'
const Login = () => import('./views/Login.vue')
const Register = () => import('./views/Register.vue')
const OperationsShell = () => import('./views/OperationsShell.vue')
const FamilyPortal = () => import('./views/FamilyPortal.vue')
const OperationsOverview = () => import('./views/OperationsOverview.vue')
const FamilyOverview = () => import('./views/FamilyOverview.vue')
const VitalsOverview = () => import('./views/VitalsOverview.vue')
const FamilyVitals = () => import('./views/FamilyVitals.vue')
const NoticeEditor = () => import('./components/NoticeEditor.vue')
const FamilyNotice = () => import('./views/FamilyNotice.vue')
const NoticeDetail = () => import('./views/NoticeDetail.vue')
const ActivityOverview = () => import('./views/ActivityOverview.vue')
const FamilyActivity = () => import('./views/FamilyActivity.vue')
const LocationOverview = () => import('./views/LocationOverview.vue')
const FamilyLocation = () => import('./views/FamilyLocation.vue')
const LifeOverview = () => import('./views/LifeOverview.vue')
const MonitoringProfiles = () => import('./views/MonitoringProfiles.vue')
const Residents = () => import('./views/Residents.vue')
const Guardians = () => import('./views/Guardians.vue')
const NoticeManagement = () => import('./views/NoticeManagement.vue')
const FamilyAccounts = () => import('./views/FamilyAccounts.vue')
const FamilyProfile = () => import('./views/FamilyProfile.vue')
const CommentManagement = () => import('./views/CommentManagement.vue')
const FamilyInvitations = () => import('./views/FamilyInvitations.vue')
const DeviceManagement = () => import('./views/DeviceManagement.vue')


const router = createRouter({
  history: createWebHashHistory(),
  routes: [
    { path: '/setup', component: () => import('./views/Setup.vue') },
	  {path: '/',
	    redirect: '/sign-in'
	  },
    {path: '/sign-in',
      component: Login
    },
    // 家属凭邀请码自助注册（无需登录）
    {path: '/join-family',
      component: Register
    },
	{path: '/family-portal',
	  component: FamilyPortal,
	  redirect: '/family-portal/overview',
	  children: [{
				  path: '/family-portal/overview',
				  meta: { title: '家属首页' }, name: 'family-portal-overview',
				  component: FamilyOverview
				},
				{
					path: '/family-portal/vitals',
					meta: { title: '心率血压' }, name: 'family-portal-vitals',
					component: FamilyVitals
				},
				{
					path: '/family-portal/activity',
					meta: { title: '睡眠步数' }, name: 'family-portal-activity',
					component: FamilyActivity
				},
				{
					path: '/family-portal/location',
					meta: { title: '位置信息' }, name: 'family-portal-location',
					component: FamilyLocation
				},
				{
					path: '/family-portal/profile',
					meta: { title: '个人中心' }, name: 'family-portal-profile',
					component: FamilyProfile
				},
				{
					path: '/family-portal/notice',
					meta: { title: '公告内容' }, name: 'family-portal-notice',
					component: FamilyNotice
				}]
	},
    { path: '/operations',
      component: OperationsShell,
      redirect: '/operations/overview',
      children: [
				{
				  path: '/operations/overview',
				  meta: { title: '运营总览' }, name: 'operations-overview',
				  component: OperationsOverview
				},
				{
					path: '/operations/vitals',
					meta: { title: '心率与血压' }, name: 'operations-vitals',
					component: VitalsOverview
				},
				{
					path: '/operations/activity',
					meta: { title: '睡眠与步数' }, name: 'operations-activity',
					component: ActivityOverview
				},
				{
					path: '/operations/locations',
					meta: { title: '位置与告警' }, name: 'operations-locations',
					component: LocationOverview
				},
				{
					path: '/operations/life',
					meta: { title: '生活信息' }, name: 'operations-life',
					component: LifeOverview
				},
				{
					path: '/operations/monitoring-profiles',
					meta: { title: '监测阈值档案' }, name: 'operations-monitoring-profiles',
					component: MonitoringProfiles
				},
				{
					path: '/operations/residents',
					meta: { title: '老人档案' }, name: 'operations-residents',
					component: Residents
				},
				{
					path: '/operations/guardians',
					meta: { title: '监护人信息' }, name: 'operations-guardians',
					component: Guardians
				},
				{
					path: '/operations/notices/edit',
					meta: { title: '编辑公告' }, name: 'operations-notices-edit',
					component: NoticeEditor, props: { editing: true }
				},
				{
					path: '/operations/notices/new',
					meta: { title: '发布公告' }, name: 'operations-notices-new',
					component: NoticeEditor, props: { editing: false }
				},
				{
					path: '/operations/notices/detail',
					meta: { title: '公告内容' }, name: 'operations-notices-detail',
					component: NoticeDetail
				},
				{
					path: '/operations/notices',
					meta: { title: '公告管理' }, name: 'operations-notices',
					component: NoticeManagement
				},

				{
				  path: '/operations/families',
				  meta: { title: '家属账号' }, name: 'operations-families',
				  component: FamilyAccounts
				},
				{
				  path: '/operations/comments',
				  meta: { title: '评论管理' }, name: 'operations-comments',
				  component: CommentManagement
				},
				{
				  path: '/operations/invitations',
				  meta: { title: '家属注册邀请码' }, name: 'operations-invitations',
				  component: FamilyInvitations
				},
				{
				  path: '/operations/devices',
				  meta: { title: '设备管理' }, name: 'operations-devices',
				  component: DeviceManagement
				}
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
