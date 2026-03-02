import { createRouter, createWebHashHistory } from 'vue-router'

const routes = [
    // 控制台 — 默认页，未部署时显示部署流程，已部署时显示仪表盘
    {
        path: '/',
        redirect: '/console',
    },
    {
        path: '/console',
        name: 'Console',
        component: () => import('@/views/Console.vue'),
    },
    // 部署流程子页面
    {
        path: '/setup',
        name: 'DeploySetup',
        component: () => import('@/views/DeploySetup.vue'),
    },
    {
        path: '/progress',
        name: 'DeployProgress',
        component: () => import('@/views/DeployProgress.vue'),
    },
    // 功能菜单页面
    {
        path: '/abilities',
        name: 'Abilities',
        component: () => import('@/views/Abilities.vue'),
    },
    {
        path: '/cron',
        name: 'Cron',
        component: () => import('@/views/Cron.vue'),
    },
    {
        path: '/chat',
        name: 'Chat',
        component: () => import('@/views/Chat.vue'),
    },
    {
        path: '/conversation',
        name: 'Conversation',
        component: () => import('@/views/Conversation.vue'),
    },
    {
        path: '/agents',
        name: 'Agents',
        component: () => import('@/views/Agents.vue'),
    },
    {
        path: '/models',
        name: 'Models',
        component: () => import('@/views/Models.vue'),
    },

]

const router = createRouter({
    history: createWebHashHistory(),
    routes,
})

export default router
