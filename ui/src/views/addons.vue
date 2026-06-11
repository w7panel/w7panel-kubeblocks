<template>
    <div class="padding-20" style="height:100%;overflow:auto;">
        <a-table :data="insList" class="cptable" :pagination="false" :bordered="false">
            <template #columns>
                <a-table-column title="已安装扩展">
                    <template #cell="{ record }">{{ record.name }}</template>
                </a-table-column>
                <a-table-column title="版本" :width="300">
                    <template #cell="{ record }">{{ record.version }}</template>
                </a-table-column>
                <a-table-column title="操作" :width="200">
                    <template #cell="{ record }">
                        <span class="c-blue cursor" @click="toCreate(record)">创建</span>
                    </template>
                </a-table-column>
            </template>
        </a-table>
        <a-table :data="unInsList" class="cptable mt-20" :pagination="false" :bordered="false">
            <template #columns>
                <a-table-column title="未安装扩展">
                    <template #cell="{ record }">{{ record['addon-name'] }}</template>
                </a-table-column>
                <a-table-column title="版本" :width="300">
                    <template #cell="{ record }">1.0.1</template>
                </a-table-column>
                <a-table-column title="操作" :width="200">
                    <template #cell="{ record }">
                        <span class="c-blue cursor" @click="toInstall(record)">安装</span>
                    </template>
                </a-table-column>
            </template>
        </a-table>

        <a-drawer :width="1200" :title="schema.name||'创建'" :visible="schema.show" @ok="submitCreate" @cancel="schema.show=false;" unmountOnClose>
            <JsonForm ref="jsonform" v-if="schema.json" :data="schema.json"></JsonForm>
        </a-drawer>
    </div>
</template>
<script>
import axios from 'axios';

import {k8sproxy} from '@/utils/k8sproxy'
import JsonForm from '@/components/JsonForm.vue';

export default{
    data(){
        return {
            namespace: 'default',
            unInsList: [],
            schema: {
                show: false,
                title: '',
                json: null,
            },

            urlpre: '',
        }
    },
    created(){
        this.urlpre = `/panel-api/v1/microapp/${window.$wujie?.props?.appgroup||'kb-test'}/proxy`
        this.getList();
    },
    components: {
        JsonForm,
    },
    methods: {
        async getList(){
            await axios.get(`${this.urlpre}/api/v1/addons`).then(res=>{
                let list = res?.data?.addons;
                this.unInsList = list || [];
            })
            await k8sproxy.get('/apis/extensions.kubeblocks.io/v1alpha1/addons').then(res=>{
                let list = res?.data?.items;
                this.unInsList = this.unInsList.filter(i=>{
                    let ls = list.map(i=>i.metadata.name);
                    return ls.includes(i?.['addon-name'])? false : true;
                })
                this.insList = list.map(i=>{
                    return {
                        name: i.metadata.name,
                        version: i.spec.version,
                    }
                })
            })
        },
        toCreate(row){
            axios.get(`${this.urlpre}/api/v1/cluster/${row.name}/${row.version}`,{loading:true}).then(res=>{
                this.schema = {
                    show: true,
                    name: row.name,
                    json: res.data,
                    version: row.version,
                }
                console.log(this.schema.json)
            })
        },
        submitCreate(){
            let values = this.$refs.jsonform.getValue();
            axios.post(`${this.urlpre}/api/v1/cluster/manifest`,{
                "release_name": this.schema.name,
                "chart_ref": this.schema.name + '-cluster',
                "chart_version": this.schema.version,
                "values": values
            },{loading:true}).then(res=>{
                axios.post('/panel-api/v1/yaml?namespace='+this.namespace,res.data,{loading:true}).then(res=>{
                    if(res?.data){
                        this.$message.success('创建成功');
                        this.getList();
                        this.schema.show = false;
                    }
                })
            })
        },
        toInstall(row){
            axios.get(`${this.urlpre}/api/v1/addons/${row['addon-name']}/1.0.1.yaml`).then(res=>{
                // console.log(res.data)
                axios.post('/panel-api/v1/yaml?namespace='+this.namespace,res.data,{loading:true}).then(res=>{
                    if(res?.data){
                        this.$message.success('创建成功');
                        this.getList();
                    }
                })
            })
        },
    }
}
</script>
<style scoped>
</style>
