import axios from 'axios';

const K8S_PROXY_PREFIX = '/k8s-proxy';
const PANEL_API_PREFIX = '/panel-api/v1';

export const k8sproxy = {
    get: (path, config) => axios.get(`${K8S_PROXY_PREFIX}${path}`, config),
    post: (path, data, config) => axios.post(`${K8S_PROXY_PREFIX}${path}`, data, config),
    patch: (path, data, config) => axios.patch(`${K8S_PROXY_PREFIX}${path}`, data, config),
    put: (path, data, config) => axios.put(`${K8S_PROXY_PREFIX}${path}`, data, config),
    delete: (path, config) => axios.delete(`${K8S_PROXY_PREFIX}${path}`, config),
};