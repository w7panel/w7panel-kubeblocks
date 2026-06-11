<template>
    <div class="padding-20">
        <div v-if="needInstall.show" class="mt-40">
            <a-empty>{{needInstall.release}}应用未安装，<span class="c-blue cursor" @click="toInstall">点击安装</span></a-empty>
        </div>
        <div v-else >
            <div class="mb-20">
                <a-button type="primary" @click="createdb.show=true;"><template #icon><icon-plus /></template>新建</a-button>
            </div>
            <div>
                <a-table :data="list" class="cptable" :pagination="false" :bordered="false">
                    <template #columns>
                        <a-table-column title="名字">
                            <template #cell="{ record }">
                                <div class="df df-c">
                                    <span class="c-blue cursor" @click="$router.push('/dbdetail/'+record.name)">{{record.name}}</span>
                                    <span class="mt-4 list-title cursor" @click="editTitle(record)">
                                        <span>{{record.title||'未命名'}}</span>
                                        <icon-edit class="editbtn ml-4 c-blue cursor" />
                                    </span>
                                </div>
                            </template>
                        </a-table-column>
                        <a-table-column title="数据库版本">
                            <template #cell="{ record }">{{record.version}}</template>
                        </a-table-column>
                        <a-table-column title="状态">
                            <template #cell="{ record }">
                                <a-popover position="bottom" @popup-visible-change="v=>v?getEvents(record):null">
                                    <span v-if="record.status=='Failed'" class="c-red cursor">{{record.status}}</span>
                                    <span v-else-if="record.status=='Running'" class="c-green cursor">{{record.status}}</span>
                                    <span v-else class="c-99 cursor">{{record.status}}</span>
                                    <template #content>
                                        <div style="max-height:400px; max-width:800px; overflow:auto;">
                                            <table class="com-table"><tbody>
                                                <tr>
                                                    <td>级别</td>
                                                    <td>内容</td>
                                                    <td>详细描述</td>
                                                    <td>时间</td>
                                                </tr>
                                                <tr v-for="(record,index) in eventLs" :key="index">
                                                    <td>{{record.type}}</td>
                                                    <td>{{record.reason}}</td>
                                                    <td>{{record.message}}</td>
                                                    <td>{{record.eventTime}}</td>
                                                </tr>
                                                <tr v-if="!eventLs||!eventLs.length">
                                                    <td colspan="5" class="txt-c c-cc">没有数据</td>
                                                </tr>
                                            </tbody></table>
                                        </div>
                                    </template>
                                </a-popover>
                            </template>
                        </a-table-column>
                        <a-table-column title="配置：(CPU/内存/硬盘)">
                            <template #cell="{ record }">{{record.cpu}} / {{record.memory}} / {{record.storage}}</template>
                        </a-table-column>
                        <a-table-column title="操作">
                            <template #cell="{ record }">
                                <a-tooltip content="yaml">
                                    <i class="opt-icon" @click="openYaml(record.name)"><icon-code /></i>
                                </a-tooltip>
                                <a-popconfirm v-if="record.status!='Deleting'" :content="'确认要删除吗'" @ok="delItem(record)" position="lt">
                                    <a-tooltip content="删除">
                                        <i class="opt-icon"><icon-delete /></i>
                                    </a-tooltip>
                                </a-popconfirm>
                            </template>
                        </a-table-column>
                    </template>
                </a-table>
            </div>
        </div>

        <a-modal :visible="titleForm.show" title="修改名称" @cancel="titleForm.show=false;" @ok="submitEditTitle">
            <a-form v-model="titleForm" auto-label-width >
                <a-form-item label="名称" field="memory">
                    <a-input v-model="titleForm.title" size="large" :spellcheck="false" placeholder="请输入"  ></a-input>
                </a-form-item>
            </a-form>
        </a-modal>

        <new-db :show="createdb.show" @close="v=>{createdb.show=false;v&&getList()}"></new-db>

        <!-- yaml -->
        <yaml-drawer :show="yamlData.show" :title="yamlData.title" :data="yamlData.data" @submit="yamlData.submit" @cancel="yamlData.show=false;"></yaml-drawer>

        <a-modal title="安装详情" width="900px" v-model:visible="log.show" :footer="false" @close="testInstall();getList();">
            <div class="log-terminal" ref="bislog"></div>
        </a-modal>
    </div>
</template>
<script>
import axios from 'axios';
import {k8sproxy} from '@/utils/k8sproxy'
import newDb from '@/views/form.vue';
import dayjs from 'dayjs'
import yamlDrawer from '@/components/yaml-drawer.vue'


import { Terminal } from '@xterm/xterm';
import '@xterm/xterm/css/xterm.css';
import { FitAddon } from '@xterm/addon-fit';

export default{
    data(){
        return {
            namespaceActive: 'default',
            list: [],
            eventLs: [],
            titleForm: {},
            createdb: {},
            yamlData: {
                show: false,
                data: {},
                title: "",
                submit: ()=>{},
            },
            urlpre: '',

            needInstall: {
                show: false,
                namespace: '',
                release: '',
            },
            
            log: {
                show: false,
                data: '',
            },
            term: null,
            fitAddon: null,
        }
    },
    components: {
        newDb,
        yamlDrawer,
    },
    created(){
        this.urlpre = `/panel-api/v1/microapp/${window.$wujie?.props?.appgroup||'kb-test'}/proxy`;
        // 测试
        this.testInstall();
        this.getList();
    },
    onUnmounted() {
        try {
            if(this.term){ this.term.dispose(); this.term = null; this.fitAddon = null; }
        } catch{}
    },

    methods: {
        async testInstall(){
            const {data} = await axios.get(`${this.urlpre}/api/v1/kubeblocks/release/check`).catch(()=>({}))
            if(!data?.exists){
                this.needInstall.show = true;
                this.needInstall.release = data.release;
                this.needInstall.namespace = data.namespace;
            }
        },
        toInstall(){
            this.log.data = '';
            this.openLog();
            axios.get(`${this.urlpre}/api/v1/install/${this.needInstall.release}`,{
                timeout: 0,
                // 👇 关键：开启流式响应，实时获取返回数据
                responseType: 'stream',
                onDownloadProgress: async (progressEvent) => {
                    try {
                        // 实时获取返回的文本内容
                        const chunk = progressEvent?.event?.target?.responseText || progressEvent?.target?.responseText;
                        if (chunk) {
                            this.log.data += chunk;

                            this.writeChunk(chunk);
                        }
                    } catch (e) {}
                },
            }).then(()=>{}).catch(()=>{})
            
        },
        getList(){
            k8sproxy.get('/apis/apps.kubeblocks.io/v1/namespaces/'+ this.namespaceActive +'/clusters').then(res=>{
                let list = res?.data?.items;
                list = list.map(i=>{
                    let events = i?.status?.conditions?.map(c=>({
                            type: c.type,
                            reason: c.reason,
                            message: c.message,
                            eventTime: dayjs(c.lastTransitionTime).format('YYYY-MM-DD HH:mm:ss'),
                    })) || [];
                    let cs = i.spec?.componentSpecs?.[0] || {};
                    return {
                        name: i.metadata?.name,
                        title: i.metadata?.annotations?.title,
                        version: cs?.serviceVersion,
                        status: i.status?.phase,
                        cpu: cs?.resources?.limits?.cpu || '',
                        memory: cs?.resources?.limits?.memory || '',
                        storage: cs?.volumeClaimTemplates?.[0]?.spec?.resources?.requests?.storage,
                        events: events,
                    }
                })
                this.list = list;
            })
        },
        getEvents(row){
            this.eventLs = row.events;
        },
        editTitle(row){
            this.titleForm = {
                show: true,
                title: row.title || '',
                name: row.name,
            }
        },
        // 删除
        delItem(row){
            k8sproxy.delete('/apis/apps.kubeblocks.io/v1/namespaces/'+ this.namespaceActive +'/clusters/'+row.name).then(res=>{
                this.$message.success('操作成功');
                this.getList();
            })
        },
        submitEditTitle(){
            k8sproxy.patch('/apis/apps.kubeblocks.io/v1/namespaces/'+ this.namespaceActive +'/clusters/'+this.titleForm.name, {
                metadata:{
                    annotations: {
                        title: this.titleForm.title,
                    }
                }
            },{headers: {'Content-Type': 'application/merge-patch+json'}}).then(res=>{
                this.titleForm.show = false;
                this.$message.success('操作成功');
                this.getList();
            });
        },
        openYaml(name){
            k8sproxy.get("/apis/apps.kubeblocks.io/v1/namespaces/"+this.namespaceActive+"/clusters/"+name).then(res=>{
                if(!res?.data){return}
                this.yamlData = {
                    show: true,
                    data: res?.data,
                    title: res?.data?.metadata?.annotations?.title || res?.data?.metadata?.name,
                    submit: (data)=>{
                        return k8sproxy.put("/apis/apps.kubeblocks.io/v1/namespaces/"+this.namespaceActive+"/clusters/" + data?.metadata?.name, data).then(res=>{
                            this.$message.success("修改成功");
                            this.yamlData = {...this.yamlData, show:false,};
                        })
                    }
                }
            })
        },
        
        openLog(){
            this.log.show = true;
            this.$nextTick(()=>{
                this.initTerm();
                if(this.log.data){
                    let text = this.log.data.replace(/\x20+/g, ' ');
                    text = text.replace(/(?<!\r)\n/g, '\r\n');
                    this.$nextTick(()=>{
                        this.term?.write(text);
                        this.fitAddon?.fit();
                    });
                }
            });
        },
        initTerm(){
            if(this.term){
                try { this.term.dispose(); } catch{}
                this.term = null;
                this.fitAddon = null;
            }
            let dom = this.$refs.bislog;
            if(!dom) return;
            dom.innerHTML = '';
            this.term = new Terminal({
                cursorBlink: false,
            });
            this.term.open(dom);
            this.fitAddon = new FitAddon();
            this.term.loadAddon(this.fitAddon);
            this.$nextTick(()=>{
                this.fitAddon.fit();
            });
        },
        writeChunk(chunk){
            if(!this.term) return;
            try {
                let text = chunk.replace(/\x20+/g, ' ');
                text = text.replace(/(?<!\r)\n/g, '\r\n');
                this.term.write(text);
            } catch{}
        },
    }
}
</script>
<style scoped>
.list-title .editbtn{display:none;}
.list-title:hover .editbtn{display:inline-block;}

.log-terminal {
    width: 100%;
    flex: 1;
    min-height: 0;
    border: 1px solid var(--color-neutral-3);
}
</style>