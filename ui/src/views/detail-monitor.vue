<template>
    <div class="bg-white padding-20 " style="height:100%;">
        <div class="df" style="height:100%;">
            <div style="width:50%;padding-right:10px;">
                <pods-charts :list="list" type="cpu"></pods-charts>
            </div>
            <div style="width:50%;padding-left:10px;">
                <pods-charts :list="list" type="memory"></pods-charts>
            </div>
        </div>
    </div>
</template>

<script>
import podsCharts from '@/components/pods-charts.vue'
// import { useNamespaceStore } from '@/store';
// import axios from 'axios'
import {k8sproxy as axios} from '@/utils/k8sproxy'

export default {
    props: ['data'],
    data(){
        return {
            namespaceActive: 'default',
            list: [],
        }
    },
    components: {
        podsCharts,
    },
    created(){
        // this.namespaceActive = useNamespaceStore().namespace;
        this.getList();
    },
    watch:{
        data(v){
            if(!v){return}
            this.getList();
        },
    },
    methods: {
        getList(){
            if(!this.data || !Object.keys(this.data).length){return}

            let labelSelector = 'labelSelector=app.kubernetes.io/instance=' + this.data.metadata.name;
            axios.get("/api/v1/namespaces/"+ this.namespaceActive +"/pods?"+labelSelector,{
                loading:true
            }).then(res=>{
                let items = res?.data?.items || [];
                this.list = items.map(item=>item?.metadata?.name);
            });
        },
    },

}
</script>

<style>

</style>