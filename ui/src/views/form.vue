<template>
    
    <a-drawer :width="900" :visible="visible" @ok="submit" @cancel="closeDrawer()" unmountOnClose>
        <template #title>集群数据库</template>
        <div class="padding-10">
            <a-form ref="form" :model="form" :rules="rules" validate-trigger="blur" auto-label-width>
                <a-form-item label="类型" field="type">
                    <a-radio-group v-model="form.type" @change="v=>form.version=versions[v][0]">
                        <a-radio value="postgresql">PostgreSQL</a-radio>
                        <a-radio value="mongodb">MongoDB</a-radio>
                        <a-radio value="mysql">MySQL</a-radio>
                        <a-radio value="redis">Redis</a-radio>
                    </a-radio-group>
                </a-form-item>
                <a-form-item label="版本" field="version">
                    <a-select v-model="form.version" placeholder="请选择版本">
                        <a-option v-for="(item,index) in versions[form.type]" :value="item" :key="index">{{ item }}</a-option>
                    </a-select>
                </a-form-item>
                <a-form-item label="名字" field="name">
                    <a-input v-model="form.name" placeholder="请输入"></a-input>
                </a-form-item>
                <a-form-item label="CPU">
                    <a-slider v-model="form.cpu" :min="0" :max="8" :marks="cpuMarks" :format-tooltip="v=>cpuMarks[v]"/>
                </a-form-item>
                <a-form-item label="内存">
                    <a-slider v-model="form.memory" :min="0" :max="8" :marks="memoryMarks" :format-tooltip="v=>memoryMarks[v]" />
                </a-form-item>
                <a-form-item label="实例数" field="instance">
                    <a-input-number v-model="form.instance" mode="button" />
                </a-form-item>
                <a-form-item label="存储" field="storageclass">
                    <a-select v-model="form.storageclass" placeholder="请选择">
                        <a-option v-for="item in storageClasses" :key="item" :label="item.name" :value="item.name"></a-option>
                    </a-select>
                </a-form-item>
                <a-form-item label="存储大小" field="storage">
                    <a-input-number v-model="form.storage" field="storage">
                        <template #suffix>Gi</template>
                    </a-input-number>
                </a-form-item>
                <a-form-item label="多写">
                    <a-switch v-model="form.accessModes" checked-value="ReadWriteMany" unchecked-value="ReadWriteOnce"></a-switch>
                </a-form-item>
            </a-form>
        </div>
    </a-drawer>
</template>
<script>
// import axios from 'axios';
import {k8sproxy as axios} from '@/utils/k8sproxy'

export default{
    props: ['show'],
    data(){
        return {
            namespaceActive: 'default',
            rules: {
                type: [{ required: true, message: '内容不能为空', trigger: 'blur' }],
                version: [{ required: true, message: '内容不能为空', trigger: 'blur' }],
                name: [{ required: true, message: '内容不能为空', trigger: 'blur' }],
                instance: [{ required: true, message: '内容不能为空', trigger: 'blur' }],
                storageclass: [{ required: true, message: '内容不能为空', trigger: 'blur' }],
                storage: [{ required: true, message: '内容不能为空', trigger: 'blur' }],
            },
            cpuMarks: ['0.5','1','2','3','4','5','6','7','8'],
            memoryMarks: ['0.5Gi','1Gi','2Gi','4Gi','6Gi','8Gi','12Gi','16Gi','32Gi'],
            versions: {
                postgresql: ['12.14.0','12.14.1','12.15.0','12.22.0','14.7.2','14.8.0','14.18.0','15.7.0','15.13.0','16.4.0','16.9.0','17.5.0',],
                mongodb: ['8.0.8','7.0.18','6.0.21','5.0.29','4.4.29',],
                mysql: ['5.7.44', '8.0.35', '8.4.0', '8.4.1', '8.4.2',],
                redis: ['6.2.13', '6.2.14', '6.2.16','7.0.15', '7.0.16', '7.2.4', '7.4.0','8.0.4', '8.0.6', '8.2.0'],
            },
            form: {
                type: 'postgresql',
                version: '',
                name: '',
                cpu: 0,
                memory: 0,
                instance: 1,
                storage: 20,
                storageclass: "",
                accessModes: "ReadWriteOnce",
            },
            storageClasses: [],
        }
    },
    created(){
        this.getStorageClasses();
    },
    watch: {
        show(v){
            this.visible = v;
            v && this.init();
        },
    },
    methods: {
        getStorageClasses(){
            axios.get('/apis/storage.k8s.io/v1/storageclasses').then(res=>{
                let list = res?.data?.items || [];
                list = list.map(i=>{
                    return {
                        name: i.metadata.name,
                    }
                })
                this.storageClasses = list;
            })
        },
        init(){
            this.form = {
                type: 'postgresql',
                version: this.versions['postgresql'][0],
                name: '',
                cpu: 0,
                memory: 0,
                instance: 1,
                storage: 20,
                storageclass: "",
                accessModes: "ReadWriteOnce",
            }
        },
        submit(){
            
            this.$refs.form.validate((err) => {
                if (err) {
                    this.$refs.form.scrollToField(Object.keys(err)[0])
                    return;
                }
                
                let f = this.form;
                let topology = {
                    mongodb: 'replicaset',
                    mysql: '',
                    postgresql: 'replication',
                    redis: 'replication',
                }
                let json = {
                    "apiVersion": "apps.kubeblocks.io/v1",
                    "kind": "Cluster",
                    "metadata": {
                        "name": f.name,
                        "namespace": this.namespaceActive,
                    },
                    "spec": {
                        "terminationPolicy": "Delete",
                        "clusterDef": f.type,
                        ...(topology[f.type]? {
                            "topology": topology[f.type],
                        }:{}),
                        "componentSpecs": [
                            {
                                "name": f.type,
                                "serviceVersion": f.version,
                                ...(f.type=='mongodb'?{}:{
                                    "disableExporter": false,
                                }),
                                "replicas": f.instance,
                                "resources": {
                                    "limits": {
                                        "cpu": this.cpuMarks[f.cpu],
                                        "memory": this.memoryMarks[f.memory],
                                    },
                                    "requests": {
                                        "cpu": this.cpuMarks[f.cpu],
                                        "memory": this.memoryMarks[f.memory],
                                    },
                                },
                                "volumeClaimTemplates": [
                                    {
                                        "name": 'data',
                                        "spec": {
                                            "storageClassName": f.storageclass,
                                            "accessModes": [f.accessModes],
                                            "resources": {
                                                "requests": {
                                                    "storage": f.storage + 'Gi'
                                                }
                                            }
                                        }
                                    },
                                ]
                            }
                        ]
                    },
                }
    
                if(f.type=='redis'){
                    let obj = JSON.parse(JSON.stringify(json.spec.componentSpecs[0]));
                    obj.name = 'redis-sentinel';
                    json.spec.componentSpecs[1] = obj;
                }
                console.log(json);
                
                
                axios.post('/apis/apps.kubeblocks.io/v1/namespaces/'+ this.namespaceActive +'/clusters',json).then(res=>{
                    this.$message.success('操作成功');
                    this.closeDrawer(true);
                });
            });
        },
        closeDrawer(refreshList){
            this.visible = false;
            this.$emit('close',refreshList);
        },
    }
}
</script>
<style scoped>
</style>