import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    component: () => import('@/layouts/MainLayout.vue'),
    redirect: '/dashboard',
    children: [
      {
        path: 'dashboard',
        name: 'Dashboard',
        component: () => import('@/pages/dashboard/index.vue'),
        meta: {
          title: '仪表盘',
          icon: 'Odometer',
        },
      },
      {
        path: 'rules',
        name: 'Rules',
        component: () => import('@/pages/rules/index.vue'),
        meta: {
          title: '规则管理',
          icon: 'List',
        },
      },
      {
        path: 'rules/:id',
        name: 'RuleDetail',
        component: () => import('@/pages/rules/detail.vue'),
        meta: {
          title: '规则详情',
          hidden: true,
        },
      },
      {
        path: 'events',
        name: 'Events',
        component: () => import('@/pages/events/index.vue'),
        meta: {
          title: '事件日志',
          icon: 'Document',
        },
      },
      {
        path: 'settings',
        name: 'Settings',
        component: () => import('@/pages/settings/index.vue'),
        meta: {
          title: '系统设置',
          icon: 'Setting',
        },
      },
    ],
  },
  {
    path: '/:pathMatch(.*)*',
    name: 'NotFound',
    component: () => import('@/pages/error/404.vue'),
  },
]

const router = createRouter({
  history: createWebHistory('/ai-plugin/'),
  routes,
})

// Update page title
router.beforeEach((to, _from, next) => {
  const title = to.meta.title as string
  document.title = title ? `${title} - AI 协同工作规范` : 'AI 协同工作规范'
  next()
})

export default router
