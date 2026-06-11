<template>
    <div>
        <a-drawer :width="800" :visible="showDrower" @cancel="cancel" unmountOnClose :footer="false" :popup-container="false?'#allmodalbox':'body'">
            <template #title>{{title}}</template>
            <yaml-editor v-if="showDrower" :yaml="yaml" :disabled="disabled" :nofooter="nofooter" @cancel="cancel" @submit="submit"></yaml-editor>
        </a-drawer>
    </div>
</template>

<script>
import yamlEditor from "@/components/yaml-editor.vue";
import jsyaml from "js-yaml";

export default {
    props: ['title', 'data', 'show','returnYaml','disabled', 'nofooter'],
    components: {
        yamlEditor,
    },
    data() {
        return {
            yaml: "",
            showDrower: false,
        }
    },
    created(){
        if(this.show){
            if(typeof this.data == 'string'){
                this.yaml = this.data;
            }else{
                let data = JSON.parse(JSON.stringify(this.data));
                if(data.metadata?.managedFields){
                    delete data.metadata?.managedFields;
                }
                delete data.metadata?.resourceVersion;
                delete data.metadata?.uid;
                this.yaml = jsyaml.dump(data, { indent: 4 });
            }
            this.showDrower = true;
        }
    },
    watch: {
        show(v){
            if(typeof this.data == 'string'){
                this.yaml = this.data;
            }else{
                let data = this.data;
                try{
                    data = JSON.parse(JSON.stringify(this.data));
                    if(data.metadata?.managedFields){
                        delete data.metadata?.managedFields;
                    }
                }catch{}
                this.yaml = jsyaml.dump(data, { indent: 4 });
            }
            this.showDrower = v;
        }
    },
    methods: {
        cancel(){
            this.$emit('cancel');
        },
        submit(yamlData){
            if(this.returnYaml){
                this.$emit('submit',yamlData);
                return;
            }
            let data = jsyaml.load(yamlData);
            
            delete data.metadata.resourceVersion
            delete data.metadata.uid
            delete data.metadata.creationTimestamp
            
            this.$emit('submit', data);
        }
    },
}
</script>

<style>
.yaml-drawer .el-drawer__header{margin-bottom:0!important; padding:10px 20px!important;}
.yaml-drawer .el-drawer__body{padding:10px 20px 20px;}
</style>