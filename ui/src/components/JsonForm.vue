<template>
    <a-form v-if="form" auto-label-width>
        <FormItem
            v-for="(value,key) in formJson.properties"
            v-model="form[key]"
            :keys="key"
            :data="value"
            :name="key"
            :storages="storages"
        ></FormItem>
    </a-form>
</template>
<script>
import formJson from '@/assets/values.schema.json';
import FormItem from '@/components/FormItem.vue';
import axios from 'axios';
import {k8sproxy} from '@/utils/k8sproxy'

export default{
    props: ['data'],
    data(){
        return {
            formJson: {},
            form: null,
            storages: [],
        }
    },
    created(){
        this.formJson = this.data;
        this.form = this.jsonToForm(this.data);
        this.getStorage();
    },
    components: {
        FormItem,
    },
    watch: {
        data(){
            this.formJson = this.data;
            this.form = this.jsonToForm(this.data);
        }
    },
    methods: {
        getStorage(){

            k8sproxy.get('/apis/storage.k8s.io/v1/storageclasses').then(res=>{
                let data = res?.data || [];
                let list = data.items || [];
                list = list.map(item=>{
                    return item.metadata.name;
                })
                list = list.filter(i=>i!='longhorn'&&i!='longhorn-static')
                this.storages = list;
            });
        },
        jsonToForm(json){
            if(!json){return null;}
            let getVal = (data)=>{
                if(data.type=='object'){
                    let o = {};
                    for(let i in data.properties){
                        o[i] = getVal(data.properties[i])
                    }
                    return o;
                }else if(data.type=='array'){
                    return [];
                }else if(data.type=='boolean'){
                    return data.default || false;
                }else{
                    return data.default || '';
                }
            }
            return getVal(json)
        },
        getValue(){
            return this.form;
            console.log(this.form)
        }
    }
}
</script>
<style scoped>
</style>