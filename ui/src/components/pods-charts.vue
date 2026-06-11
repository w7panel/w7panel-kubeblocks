<template>
    <div class="df df-c" style="height: 100%">
        <div v-if="type">
            <div v-if="noTitle!==true" class="fs-16 b" style="margin-bottom: 40px">{{titles[type]}}</div>
        </div>
        <a-tabs v-else v-model="activeType" class="df-s0" @change="(v) => (activeType = v)">
            <a-tab-pane key="cpu" title="CPU"> </a-tab-pane>
            <a-tab-pane key="memory" title="内存"> </a-tab-pane>
        </a-tabs>
        <div v-if="activeType == 'cpu'" class="fc chartbox">
            <!-- <div class="title b">CPU</div> -->
            <div id="cpuchart" class="chart"></div>
        </div>
        <div v-else class="fc chartbox">
            <!-- <div class="title b">内存</div> -->
            <div id="memorychart" class="chart"></div>
        </div>
    </div>
</template>

<script>
// import axios from 'axios';
import * as echarts from 'echarts';
import CryptoJS  from 'crypto-js';
import dayjs from 'dayjs';
import {k8sproxy as axios} from '@/utils/k8sproxy'

export default {
    props: ['list', 'type','noTitle', 'pickerValue', 'step'],
    data() {
        return {
            titles: {
                cpu: 'CPU',
                memory: '内存',
            },
            dark: false,
            namespaceActive: '',
            activeType: 'cpu',
            timeType: 'minute',
            timeRange: [],

            menuIndex: 'cpu',
            chartType: 'cpu',

            cpu: {
                loading: false,
                x: [],
                y: [],
            },
            memory: {
                loading: false,
                x: [],
                y: [],
            },
            userInfo: {},
        };
    },
    created() {
        if (this.type) {
            this.activeType = this.type;
        }
    },
    
    mounted() {
        this.init();
    },
    beforeDestroy() {
        window.removeEventListener('resize', this.resize);
    },
    watch: {
        'list'(v) {
            this.init();
        },
        'activeType'(v) {
            this.chartInit(v);
        },
        'dark.isDark'(v) {
            this.chartInit(this.activeType);
        },
        pickerValue(){
            this.chartInit(this.activeType);
        },
        step(){
            this.chartInit(this.activeType);
        },
    },
    methods: {
        // translateToHostName(name, namespace, clusterName) {
        //     const namePrefix = `${name}-${namespace}-${clusterName}`;
        //     const nameKey = `${name}+${namespace}+${clusterName}`;
        //     const nameSuffix = CryptoJS.enc.Hex.stringify(CryptoJS.enc.Utf8.parse(nameKey));
        //     const fullPath = `${namePrefix}-${nameSuffix}`;

        //     if (fullPath.length < 64) return fullPath;

        //     const hash = CryptoJS.SHA256(fullPath).toString(CryptoJS.enc.Hex);
        //     const validChar = /[a-z0-9]/.test(fullPath[56]) 
        //         ? fullPath.substring(0, 57) + "-" + hash.substring(0, 5)
        //         : fullPath.substring(0, 56) + "-" + hash.substring(0, 6);

        //     return validChar;
        // },
        resize() {},
        init() {
            if (!this.list?.length) { return; }
            this.$nextTick(() => {
                this.chartInit(this.activeType);
            });
        },
        getChart(Name, MetricName) {
            let time = parseInt(Date.now()/1000);
            let startTime = time - 3600;
            let endTime = time;
            let step = 15;

            if(this.pickerValue?.length==2){
                startTime = dayjs(this.pickerValue[0]).unix();
                endTime = dayjs(this.pickerValue[1]).unix();
            }
            if(this.step){
                step = this.step;
            }

            let agent = '/api/v1/namespaces/default/services/vmsingle-w7panel-metrics-k8s-offline-metrics-single:8429/proxy';

            // if(this.userInfo?.["k3k.io/cluster-mode"]=='virtual'){
            //     agent = '/api/v1/namespaces/default/services/vmsingle-w7panel-metrics-k8s-offline-metrics-single:8429/proxy';
            // }
            // if(this.userInfo?.["k3k.io/cluster-mode"]=='shared'){
            //     agent = '/k8s/v1/namespaces/default/services/vmsingle-w7panel-metrics-k8s-offline-metrics-single:8429/proxy-root'
            // }
            let path = agent + '/prometheus/api/v1/query_range';
            return axios.get(path,{
                params: {
                    query:
                        MetricName == 'cpu'
                        ? 'rate(pod_cpu_usage_seconds_total{pod="' + Name + '"})'
                        : 'pod_memory_working_set_bytes{pod="' + Name + '"}',
                    start: startTime,
                    end: endTime,
                    step: step,
                },
            }).then((res) => {
                return res?.data;
            });
            // return axios.get(`/api/v1/dashboard/namespaces/${this.namespaceActive}/pod-list/${Name}/metrics/${MetricName}/minute`).then(res=>{
            //     return res?.data;
            // })
        },
        async chartInit(chartType) {
            let c = this[chartType];
            c.loading = true;
            // let chart = null;
            // if(chart){ chart.dispose();}
            c.x = {};
            c.y = [];
            let dw = chartType == 'cpu' ? '核' : 'M';

            let nodes = this.list;
            for (let i = 0; i < nodes?.length; i++) {
                let nodeName = nodes[i];
                // if(this.userInfo?.['k3k.io/cluster-mode']=='shared'){
                //     nodeName = this.translateToHostName(nodeName,'default', this.userInfo?.['w7.cc/k3k-name'])
                // }
                let { data } = await this.getChart(nodeName, chartType);
                if (!data?.result?.length) { continue; }
                c.x = {
                    type: 'category',
                    data: data.result[0].values?.map((item) =>dayjs(item[0] * 1000).format('YYYY-MM-DD hh:mm:ss')),
                };
                c.y.push({
                    name: nodes[i],
                    type: 'line',
                    data: data.result[0].values?.map((item) => {
                        if (chartType == 'cpu') { return (item[1] / 1).toFixed(4); }
                        if (chartType == 'memory') { return (item[1] / 1024 / 1024).toFixed(2); }
                    }),
                    smooth: true,
                });
            }

            c.loading = false;

            let option = {
                tooltip: {
                    trigger: 'axis',
                    appendToBody: this.inMicro,
                    // alwaysShowContent: true,
                },
                legend: {
                    show: true,
                    // type: 'scroll',
                    // orient: 'vertical',
                    // top: 520,
                    // height: 120,
                },
                grid: {
                    left: '3%',
                    right: '4%',
                    // bottom: 140,
                    containLabel: true,
                },
                // xAxis: [],//c.x,
                yAxis: {
                    type: 'value',
                    axisLabel: {
                        formatter: '{value} ' + dw,
                    },
                },
                // series: [],//c.y,
            };
            
            if (chartType == 'cpu') {
                let max = 1;
                c.y.forEach((item) => {
                    let maxItem = Math.max(...item.data);
                    if (maxItem > max) {
                        max = maxItem;
                    }
                });
                option.yAxis.max = Math.ceil(max);
            }

            if (c.x?.data?.length) {
                option.xAxis = c.x;
                option.series = c.y;
                let chart = null;
                if (chart) {chart.dispose();}
                let dom = document.getElementById(chartType + 'chart');
                if (dom) {
                    dom.removeAttribute('_echarts_instance_');
                    chart = echarts.init(dom);
                    this.resize = () => {
                        chart.resize();
                    };
                    window.addEventListener('resize', this.resize);
                    chart.setOption(option);
                }
            }
        },
    },
};
</script>

<style scoped>
  .chartbox {
    background: #fff;
  }
  .chartbox .title {
    padding: 16px;
  }
  .chart {
    height: 100%;
    background: #fff;
  }
</style>
