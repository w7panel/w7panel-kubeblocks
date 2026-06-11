import axios from 'axios'
import { Notification } from '@arco-design/web-vue';

import { useLoadingStore } from '@/store';
// 请求拦截器
// axios.defaults.baseURL = window?.$wujie?.props?.url || '';

axios.interceptors.request.use(
    (config) => {
        config.headers = config.headers || {};
        
        const test = `eyJhbGciOiJSUzI1NiIsImtpZCI6Im9xVkttSXBhLXI4MldWSE1Nb3lPTjNjVXdycnJJcVF1MW8tbjluanpBdmsifQ.eyJhdWQiOlsiYWRtaW4iLCJmb3VuZGVyIiwiNDkwMjA5IiwiaHR0cHM6Ly9rdWJlcm5ldGVzLmRlZmF1bHQuc3ZjLmNsdXN0ZXIubG9jYWwiLCJrM3MiXSwiZXhwIjoxNzc3MjcyNDcwLCJpYXQiOjE3NzcyNjg4NzAsImlzcyI6Imh0dHBzOi8va3ViZXJuZXRlcy5kZWZhdWx0LnN2Yy5jbHVzdGVyLmxvY2FsIiwianRpIjoiMzM0NjEzYjYtYjYwMy00MTE3LTkwMWEtMjcyNGZlNzQ0NGEwIiwia3ViZXJuZXRlcy5pbyI6eyJuYW1lc3BhY2UiOiJkZWZhdWx0Iiwic2VydmljZWFjY291bnQiOnsibmFtZSI6ImFkbWluIiwidWlkIjoiNmZiMmI2MDUtYWFiZi00Mzg4LWE0NjUtYjRmZTNkNmE0MDI0In19LCJuYmYiOjE3NzcyNjg4NzAsInN1YiI6InN5c3RlbTpzZXJ2aWNlYWNjb3VudDpkZWZhdWx0OmFkbWluIn0.wFGvybtZFmyiWFVZOb5ur_8KD-MDippQMRyWjtL07heRk7Rwvwfw-H2BqqCg2ZiL7qMibkH8TiTc4dO1Bx7cmcLsGbgOcZuspRjyuRGs0SSwBD8PrXs_xI1hW96b5MUfR52ZlYTrSmF7y3fFzioFMeJurp-7RouT2DXMzGmm80wWmwUaAny2drp-AnDffPiHpVOx2X9B8ASuse9Z6W_gEMW5MkMhv0K9J5U41euUG7XAbgit8Ocx5RwsA_mEbb4ZH6NeR6TQL5_eiTETu33XeXiji-PVz975pYk-0OSj1xv1trk4Q-ZCzf-Wr3FoqZ6o9nHtBm4ecSDfO16uB3sTuw`
        const token = window.$wujie?.props?.paneltoken || test;
        if(token){
            config.headers.Authorization = `Bearer ${token}`;
        }
        if(config?.customToken){
            config.headers.Authorization = `Bearer ${config.customToken}`;
        }
        if(config?.loading){
            useLoadingStore().loading = true;
        }
        return config
    },
    (error) => {
        if(error?.config?.loading){
            useLoadingStore().loading = false;
        }
        return Promise.reject(error)
    }
)

// 响应拦截器
axios.interceptors.response.use(
    (res) => {
        if(res?.config?.loading){
            useLoadingStore().loading = false;
        }
        if(res.status>=200 && res.status<=300){
            return res;
        }
        
        return Promise.reject(new Error(res.message || '请求失败'))
    },
    (error) => {
        if(error?.config?.loading){
            useLoadingStore().loading = false;
        }
        if (error.config.noAlert){
            return Promise.reject(error);
        }
        if(typeof error?.response?.data?.error == 'string'){
            Notification.error({title: 'Error',content: error.response.data.error,});
        }
        return Promise.reject(error);
    }
)