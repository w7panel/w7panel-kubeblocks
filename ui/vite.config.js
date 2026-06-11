import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import path from 'path'

const proxyUrl = 'http://172.16.1.162:9090';
// const proxyUrl = 'https://idc.w7.com';
// const proxyUrl = 'http://kbtest.fan.b2.sz.w7.com';

// https://vite.dev/config/
export default defineConfig({
    plugins: [vue()],
    resolve: {
        alias: {
            '@': path.resolve(__dirname, 'src')
        }
    },
    base: './',
    server: {
        proxy: {
            '/k8s/v1/namespaces/longhorn-system/services/longhorn-backend/proxy/v1': {
                target: proxyUrl,
                changeOrigin: true,
                ws: true,
            },
            '/k8s/v1/namespaces/longhorn-system/services/longhorn-backend:9500/proxy/v1': {
                target: proxyUrl,
                changeOrigin: true,
                ws: true,
            },
            '/version': {
                target: proxyUrl,
                changeOrigin: true,
                ws: true,
            },
            '/apis': {
                target: proxyUrl,
                changeOrigin: true,
                ws: true,
            },
            '/api/v1/proxy': proxyUrl,
            '/api/v1/zpk': proxyUrl,
            '/api/v1/helm/releases':proxyUrl,
            '/api/v1/cluster': 'http://kbtest.fan.b2.sz.w7.com',
            '/api/v1/addons': 'http://kbtest.fan.b2.sz.w7.com',
            '/panel-api': proxyUrl,
            '/api': {
                target: proxyUrl,
                changeOrigin: true,
                ws: true,
            },
            '/k8s': {
                target: proxyUrl,
                changeOrigin: true,
                ws: true,
            },
            '/s3bucket': {
                target: proxyUrl,
                changeOrigin: true,
                ws: true,
            },
            '/respo': {
                target: 'http://zpk.w7.cc',
                changeOrigin: true,
                ws: true,
            },
        },
    },
})
