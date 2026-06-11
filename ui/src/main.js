import { createApp } from 'vue'
import './style.css'
import App from './App.vue'
import ArcoVue from '@arco-design/web-vue';
import ArcoVueIcon from '@arco-design/web-vue/es/icon';
import '@arco-design/web-vue/dist/arco.css';
import '@/utils/request'
import router from './router'
import store from './store';

const app = createApp(App);
app.use(store)
app.use(router)
app.use(ArcoVue);
app.use(ArcoVueIcon);
app.mount('#app');
