import { createRouter, createWebHistory,createWebHashHistory } from 'vue-router'

// 定义路由规则
const routes = [
    {
        path: '/',
        name: 'list',
        component: () => import('@/views/list.vue')
    },
    {
        path: '/addons',
        name: 'addons',
        component: () => import('@/views/addons.vue')
    },
    {
        path: '/dbdetail/:id',
        name: 'database-detail',
        redirect: (route)=>{
            return {name:'database-detail-panel', params:route.params}
        },
        component: () => import('@/views/detail.vue'),
        meta: {},
        children: [{
            path: 'panel',
            name: 'database-detail-panel',
            component: () => import('@/views/detail-panel.vue'),
            meta: {}
        },{
            path: 'olog',
            name: 'database-detail-olog',
            component: () => import('@/views/detail-olog.vue'),
            meta: {}
        },{
            path: 'monitor',
            name: 'database-detail-monitor',
            component: () => import('@/views/detail-monitor.vue'),
            meta: {}
        },{
            path: 'ini',
            name: 'database-detail-ini',
            component: () => import('@/views/detail-ini-edit.vue'),
            meta: {}
        },],
    },
]

// 创建路由实例
const router = createRouter({
    history: createWebHashHistory(),
    routes,
})

export default router