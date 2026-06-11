
<template>
    <a-spin :loading="loading" :size="32" tip="加载中..." class="page-spin">
        <router-view />
    </a-spin>
</template>

<script>
import { useLoadingStore } from '@/store';

export default{
    data(){
        return {}
    },
    computed:{
        loading(){
            return useLoadingStore().loading
        }
    },
    created(){
        window.$wujie?.bus.$on("routeChange", (v)=>{
            this.$router.push(v)
        });
    },
}
</script>

<style scoped>
.page-spin{display:block;height:100%;}
</style>
