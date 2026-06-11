<template>
    <div class="padding-20 df df-c" style="height:100%;overflow:auto;">
        <a-breadcrumb style="margin-bottom:20px;">
            <a-breadcrumb-item class="cursor" @click="$router.push('/')">集群数据库</a-breadcrumb-item>
            <a-breadcrumb-item class="cursor" @click="$router.push('/dbdetail/'+$route.params.id+'/panel')">{{ $route.params.id }}</a-breadcrumb-item>
            <a-breadcrumb-item>{{ {panel:'概览',olog:'操作记录',monitor:'实时监控',ini:'参数设置'}[menukey] }}</a-breadcrumb-item>
        </a-breadcrumb>
        <div class="df fc">
            <a-tabs v-model:active-key="menukey" @tab-click="changeKey" class="db-detailmenu" position="left" hide-content>
                <a-tab-pane key="panel" title="概览"></a-tab-pane>
                <a-tab-pane key="olog" title="操作记录"></a-tab-pane>
                <a-tab-pane key="monitor" title="实时监控"></a-tab-pane>
                <a-tab-pane key="ini" title="参数设置"></a-tab-pane>
            </a-tabs>
            <div class="fc" style="overflow: auto;">
                <router-view :data="data" @refresh="getInfo" :dbtype="dbtype" />
            </div>
        </div>
    </div>
</template>
<script>
// import axios from 'axios'
import {k8sproxy as axios} from '@/utils/k8sproxy'
export default{
    data(){
        return {
            namespaceActive: 'default',
            defaultMenuKey: 'panel',
            menukey: 'panel',
            data: null,
            dbtype: '',
        }
    },
    created(){
        this.init();
        this.getInfo();
    },
    watch: {
        '$route.name'(){
            this.init();
        },
    },
    methods: {
        init(){
            let match = this.$route.name.match(/^database\-detail\-([\w\-]+)$/);
            match && (this.defaultMenuKey = this.menukey = match[1]);
        },
        getInfo(){
            axios.get('/apis/apps.kubeblocks.io/v1/namespaces/'+ this.namespaceActive +'/clusters/'+this.$route.params.id).then(res=>{
                this.data = res?.data;
                this.dbtype = this.data?.metadata?.labels?.['w7panel.kubeblocks.io/name'];
            });
        },
        changeKey(){
            if(this.menukey == this.defaultMenuKey){return}
            this.$router.push({name:'database-detail-'+this.menukey, params:this.$route.params});
        },
        
    }
}
</script>
<style scoped>
/* .db-detailmenu,
.db-detailmenu .arco-tabs-nav-tab{width:100%;} */
.db-detailmenu .arco-tabs-nav{width:100px; padding-top:10px;}

.db-tabs .arco-drawer-body{padding:0;}
.db-tabs .arco-tabs-tab.arco-tabs-tab-active{
    background:var(--color-bg-2);
}
</style>