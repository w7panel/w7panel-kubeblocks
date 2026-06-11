<template>
    <div>
        <div v-if="type=='object'">
            <div class="mt-10" style="color:var(--color-text-2);line-height:32px;">{{ title }}</div>
            <div class="fs-12" style="color:var(--color-text-3);line-height:20px;">{{ description }}</div>
            <div class="box">
                <template v-for="(item,key) in data.properties" :key="key">
                    <FormItem :data="item" :name="key" v-model="innerValue[key]" :storages="storages" />
                </template>
            </div>
        </div>
        <a-form-item v-if="type!='object'" :label="title" :help="description" style="margin-top:10px;">
            <a-input v-if="type=='string'" v-model="innerValue" placeholder="请输入"></a-input>
            <a-select v-if="type=='select'" v-model="innerValue" placeholder="请选择">
                <a-option v-for="(opt,index) in options" :key="index" :label="opt" :value="opt"></a-option>
            </a-select>
            <a-input-number v-if="type=='number'||type=='integer'" v-model="innerValue" placeholder="请输入"></a-input-number>
            <a-switch v-if="type=='boolean'" v-model="innerValue"></a-switch>
            <!-- <div v-if="type=='object'" style="flex:1;">
                <template v-for="(item,key) in data.properties" :key="key">
                    <FormItem :data="item" :name="key" v-model="innerValue[key]" />
                </template>
            </div> -->
        </a-form-item>
    </div>
</template>
<script>
export default{
    props: ['name','data','modelValue','storages'],
    emits: ['update:modelValue'],
    name: 'FormItem',
    data(){
        return {
            type: '',
            title: '',
            description: '',
            value: '',
            options: [],
        }
    },
    created(){
        this.init();
    },
    computed: {
        innerValue: {
            get() {
                return this.modelValue;
            },
            set(newVal) {
                this.$emit('update:modelValue', newVal);
            }
        },
    },
    watch: {
        data: 'init',
        storages(){
            if(this.name=='storageClassName'){
                this.options = this.storages;
            }
        },
    },
    methods: {
        init(){
            let d = this.data;
            if(!d){return}

            let type = d.type;
            if(Array.isArray(d.type)){
                type = d.type?.filter(i=>i!='null')?.[0]
            }

            if(type=='string'){
                this.type = d?.enum?.length? 'select' : 'string';
                this.options = d?.enum;
            }
            if(type=='integer' || type=='number' || type=='boolean'){
                this.type = type;
                this.value = d?.default || '';
            }
            if(type=='object'){
                this.type = 'object';
                this.value = {};
            }
            if(this.name=='storageClassName'){
                this.type = 'select';
                this.options = this.storages;
            }
            if(!type){return}

            this.title = d?.title || this.name;
            this.description = d?.description || '';

        },
        getValue(){
            if(this.type!=='object'){
                return this.value;
            }else{
                return 
            }
        },
    }
}
</script>
<style scoped>
.box{margin-top:10px;padding:10px 20px 20px 20px;border:1px solid var(--color-text-4);}
</style>