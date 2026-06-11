<template>
    <div class="bg-white padding-20 fc">
        
        <a-table :data="tableList" :pagination="false" :bordered="false" class="dbinitable">
            <template #columns>
                <a-table-column title="参数名">
                    <template #title>
                        <div class="df ai-c">
                            <span class="df-s0">参数名</span>
                            <a-input-search v-model="filterValue" class="ml-20" style="width:280px;" @search="tableFilter()" allow-clear @clear="tableFilter()" @press-enter="tableFilter()" placeholder="请输入搜索内容"/>
                        </div>
                    </template>
                    <template #cell="{ record }">
                        <div class="df ai-c">
                            <a-popover v-if="dbtype!='mongodb'" position="right">
                                <icon-minus v-if="record.dynamic" class="c-99 mr-10" />
                                <icon-refresh v-else style="color:rgb(var(--danger-5));" class="mr-10" />
                                <template #content>
                                    <span v-if="record.dynamic">动态参数，对这些参数进行修改会触发配置的动态重新加载，而无需重启进程。</span>
                                    <span v-else>静态参数，对这些参数进行修改会触发进程重启。</span>
                                </template>
                            </a-popover>

                            <a-popover v-if="record.description" position="right">
                                <span class="cursor lh-1">
                                    <span class="va-middle">{{record.name}}</span>
                                    <icon-info-circle class="ml-4 fs-16 va-middle c-99" />
                                </span>
                                <template #content>
                                    <div style="max-width:300px;">{{record.description}}</div>
                                </template>
                            </a-popover>
                            <span v-else>{{record.name}}</span>
                        </div>
                    </template>
                </a-table-column>
                <a-table-column v-if="dbtype!='mongodb'" title="参数默认值">
                    <template #cell="{ record }">
                        <span>{{record.defaultValue===''? '-' : record.defaultValue}}</span>
                    </template>
                </a-table-column>
                <a-table-column title="参数运行值" :width="dbtype=='mongodb'?600:400">
                    <template #cell="{ record }">
                        <div v-if="edit.name==record.name">
                            <a-select v-if="(record.type=='boolean'||record.type=='string')&&edit.options.length" v-model="edit.value" style="width:160px;" placeholder="请选择">
                                <a-option v-for="opt in edit.options" :key="opt" :value="opt" :label="opt.toString()"></a-option>
                            </a-select>
                            <a-input v-else-if="record.type=='number'||record.type=='integer'" v-model="edit.value" size="small" :spellcheck="false" type="number" style="width:160px;" placeholder="请输入"></a-input>
                            <a-input v-else v-model="edit.value" size="small" :spellcheck="false" style="width:160px;" placeholder="请输入"></a-input>
                            <a-button size="small" type="primary" class="ml-10" @click="editValue">确定</a-button>
                            <a-button size="small" type="outline" class="ml-10" @click="edit.name=''">取消</a-button>
                        </div>
                        <div v-else>
                            <span>{{record.currentValue===''? '-' : record.currentValue}}</span>
                            <icon-edit @click="changeEdit(record)" class="ml-10 fs-16 c-blue cursor editbtn" />
                        </div>
                    </template>
                </a-table-column>
                <a-table-column v-if="dbtype!='mongodb'" title="参数可修改值" :width="300">
                    <template #cell="{ record }">
                        <a-popover v-if="record.type=='string'" position="left">
                            <span class="one-hide">{{record.enum || '-'}}</span>
                            <template #content>
                                <div style="max-width:500px;">{{record.enum}}</div>
                            </template>
                        </a-popover>
                        <span v-else-if="record.type=='integer'&&record.minimum&&record.maximum">{{'[' + record.minimum + ' - ' + record.maximum + ']'}}</span>
                        <span v-else>-</span>
                    </template>
                </a-table-column>
            </template>
        </a-table>
    </div>
</template>

<script>
import axios from 'axios';
import {k8sproxy} from '@/utils/k8sproxy';
import { useLoadingStore } from '@/store';

export default {
    props: ['data','dbtype'],
    data(){
        return {
            urlpre: '',
            namespaceActive: 'default',
            clusterName: '',
            list: [],
            filterValue: '',
            tableList: [],

            edit: {
                name: '',
                origin: '',
                value: '',
                options: [],
            },
            componentName: '',
        }
    },
    created(){
        // this.namespaceActive = useNamespaceStore().namespace;
        this.urlpre = `/panel-api/v1/microapp/${window.$wujie?.props?.appgroup||'kb-test'}/proxy`;
        this.init();
    },
    watch:{
        data(v){
            if(!v){return}
            this.init();
        },
    },
    methods: {
        changeEdit(row){
            
            let options = [];
            
            if(this.dbtype=='mongodb'){
                if(row.type=='boolean'){ options = [true,false]; }
            }else{
                if(row.type=='string'){
                    options = row.enum || [];
                    options = options.map(i=>i.replace(/^"|"$/g,''));
                }
            }
            this.edit = {
                name: row.name,
                origin: row.value,
                value: row.value,
                options,
            }
        },
        editValue(){
            
            let data = {
                apiVersion: 'operations.kubeblocks.io/v1alpha1',
                kind: 'OpsRequest',
                metadata: {
                    name: this.createName(),
                    namespace: this.namespaceActive,
                },
                spec: {
                    clusterName: this.data.metadata.name,
                    force: false,
                    reconfigures: [{
                        componentName: this.componentName,
                        parameters: [{
                            key: this.edit.name,
                            value: this.edit.value,
                        }]
                    }],
                    preConditionDeadlineSeconds: 0,
                    type: 'Reconfiguring'
                },
            }
            
            k8sproxy.post('/apis/operations.kubeblocks.io/v1alpha1/namespaces/'+this.namespaceActive+'/opsrequests',data).then(res=>{
                this.$message.success('操作成功');
                this.getList();
                this.edit = {name:'',value:''};
            })
        },
        tableFilter(){
            this.tableList = this.list.filter(i=>{
                return i.name.toLowerCase().includes(this.filterValue?.toLowerCase());
            })
        },
        init(){
            this.clusterName = this.data?.metadata?.name;
            if(!this.clusterName){return}
            
            this.getList();
        },
        async getList(){

            useLoadingStore().loading = true;

            try{
                let {data} = await axios.get(`${this.urlpre}/api/v1/config/params?cluster=${this.clusterName}&namespace=default`);
                let list = await this.parseList(data);
                this.list = list;
            }catch{}
            
            useLoadingStore().loading = false;

            this.tableFilter();
        },
        async parseList(data){
            this.componentName = data?.components?.[0]?.componentName;
            let items = data?.components?.[0]?.templates?.[0]?.parameters || [];
            let list = items.map(i=>{
                return {
                    ...i,
                    currentValue: i?.currentValue || '',
                }
            })
            return list;
        },
        createName(length){
            let len = length || 8;
            let s = 'abcdefghijklmnopqrstuvwxyz';
            let p = '';
            for(var i=0; i<len; i++){
                p = p + s[parseInt(Math.random()*s.length)]
            }
            return p;
        },
    },
}
</script>

<style scoped>
.dbinitable .editbtn{display:none;}
.dbinitable tr:hover .editbtn{display:inline-block;}
</style>