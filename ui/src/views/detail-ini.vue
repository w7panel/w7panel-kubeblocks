<template>
    <div class="bg-white padding-20">
        <div class="df ai-c jc-s">
            <span class="mr-10">configmap</span>
            <a-select v-model="confiamapActive" style="width:400px;" @change="v=>getConfigmap(v)">
                <a-option v-for="(value,index) in source" :key="index" :label="value" :value="value"></a-option>
            </a-select>
        </div>
        <div class="mt-20">
            <div v-for="(value,key) in configmapData" :key="key" class="mb-20">
                <div>{{ key }}</div>
                <div class="mt-10">
                    <a-textarea :model-value="value" placeholder="Please enter something" auto-size/>
                </div>
            </div>
        </div>
    </div>
</template>

<script>
// import axios from 'axios';
import {k8sproxy as axios} from '@/utils/k8sproxy'

export default {
    props: ['data','dbtype'],
    data(){
        return {
            namespaceActive: 'default',
            clusterName: '',
            confiamapNames: [],
            confiamapActive: '',
            configmapData: {},
        }
    },
    created(){
        this.init();
    },
    watch:{
        data(v){
            if(!v){return}
            this.init();
        },
    },
    methods: {
        init(){
            this.clusterName = this.data?.metadata?.name;
            if(!this.clusterName){return}
            this.getConfigmapNames();
        },
        getConfigmapNames(){
            
            let labelSelector = 'labelSelector=app.kubernetes.io/instance=' + this.clusterName;
            axios.get('/apis/apps.kubeblocks.io/v1/namespaces/'+ this.namespaceActive +'/components?' + labelSelector).then(res=>{
                let names = [];
                res.data?.items?.map?.(i=>{
                    let arr = [];
                    i?.spec?.configs?.map?.(c=>{
                        let nm = c?.configMap?.name;
                        if(nm){names.push(nm)}
                    })
                })
                this.confiamapNames = names;
                this.confiamapActive = names?.[0] || '';
                this.getConfigmap(this.confiamapActive)
            })
        },
        getConfigmap(name){
            if(!name){return}
            axios.get('/api/v1/namespaces/'+ this.namespaceActive +'/configmaps/'+name).then(res=>{
                this.configmapData = res?.data?.data || {};
            })
        },
    },
}
</script>

<style scoped>
.dbinitable .editbtn{display:none;}
.dbinitable tr:hover .editbtn{display:inline-block;}
</style>