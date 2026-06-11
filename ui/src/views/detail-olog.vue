<template>
    <div class="bg-white padding-20 fc">
        <a-table :data="list" class="cptable" :pagination="false" :bordered="false">
            <template #columns>
                <a-table-column title="操作">
                    <template #cell="{ record }">{{record.operation}}</template>
                </a-table-column>
                <a-table-column title="操作时间">
                    <template #cell="{ record }">{{record.createTime}}</template>
                </a-table-column>
                <a-table-column title="状态">
                    <template #cell="{ record }">
                        <span class="c-99">
                            <span :class="{'c-red':record.status=='Failed', 'c-green':record.status=='Succeed'}">{{record.statusTxt}}</span>
                        </span>
                    </template>
                </a-table-column>
                <a-table-column title="操作">
                    <template #cell="{record}">
                        <a-tooltip content="yaml">
                            <i class="opt-icon" @click="openYaml(record.name)"><icon-code /></i>
                        </a-tooltip>
                    </template>
                </a-table-column>
            </template>
        </a-table>
        <!-- yaml -->
        <yaml-drawer :show="yamlData.show" :title="yamlData.title" :data="yamlData.data" @submit="yamlData.submit" @cancel="yamlData.show=false;"></yaml-drawer>
    </div>
</template>

<script>
// import { useNamespaceStore } from '@/store'
// import axios from 'axios';
import {k8sproxy as axios} from '@/utils/k8sproxy'
import yamlDrawer from '@/components/yaml-drawer.vue'
import dayjs from 'dayjs';
// import { getUserInfo } from '@/utils/auth';

export default {
    props: ['data'],
    data(){
        return {
            namespaceActive: 'default',
            list: [],
            yamlData: {
                show: false,
                data: {},
                title: "",
                submit: ()=>{},
            },
            // debug: false,
        }
    },
    created(){
        // this.debug = getUserInfo()?.['w7.cc/debug']=='true';
        // this.namespaceActive = useNamespaceStore().namespace;
        this.getList();
    },
    watch: {
        data(v){
            this.getList();
        },
    },
    components: { yamlDrawer },
    methods: {
        getList(){
            if(!this.data || !Object.keys(this.data)?.length){return}
            let labelSelector = 'labelSelector=app.kubernetes.io/instance=' + this.data.metadata.name;
            axios.get('/apis/operations.kubeblocks.io/v1alpha1/opsrequests?'+labelSelector).then(res=>{
                let items = res?.data?.items || [];
                let list = items?.map(i=>{
                    let n = i.metadata?.name;
                    let operation = i?.spec?.type;
                    switch(i?.spec?.type){
                        case 'HorizontalScaling': operation = '水平扩容'; break;
                        case 'VolumeExpansion': operation = '磁盘扩容'; break;
                        case 'VerticalScaling': operation = '垂直扩容'; break;
                        case 'Stop': operation = '停止'; break;
                        case 'Start': operation = '启动'; break;
                        case 'Restart': operation = '重启'; break;
                    }
                    let statusTxt = i?.status?.phase;
                    switch(i?.status?.phase){
                        case 'Succeed': statusTxt = '成功'; break;
                        case 'Failed': statusTxt = '失败'; break;
                        case 'Running': statusTxt = '执行中'; break;
                    }
                    return {
                        name: n,
                        type: i?.spec?.type,
                        operation: operation,
                        status: i?.status?.phase,
                        statusTxt: statusTxt,
                        createTime: dayjs(i.metadata?.creationTimestamp).format('YYYY-MM-DD hh:mm:ss'),
                    }
                })
                list.sort((a, b) => {
                    const dateA = new Date(a.createTime);
                    const dateB = new Date(b.createTime);
                    return dateB - dateA;
                });
                this.list = list;
            });
        },
        openYaml(name){
            axios.get("/apis/operations.kubeblocks.io/v1alpha1/namespaces/"+this.namespaceActive+"/opsrequests/"+name).then(res=>{
                if(!res?.data){return}
                this.yamlData = {
                    show: true,
                    data: res?.data,
                    title: res?.data?.metadata?.annotations?.title || res?.data?.metadata?.name,
                    submit: (data)=>{
                        return axios.put("/apis/operations.kubeblocks.io/v1alpha1/namespaces/"+this.namespaceActive+"/opsrequests/" + data?.metadata?.name, data).then(res=>{
                            this.$message.success("修改成功");
                            this.yamlData = {...this.yamlData, show:false,};
                        })
                    }
                }
            })
        },
    },
}
</script>

<style>

</style>