import { createApp } from 'vue'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
// Import custom theme after Element Plus base styles
import '@/styles/admin-theme.scss'
import zhCn from 'element-plus/es/locale/lang/zh-cn'
import App from './App.vue'
import router from './router'
import { createPinia } from 'pinia'
import * as echarts from 'echarts'

const app = createApp(App)
const pinia = createPinia()
app.use(ElementPlus, { locale: zhCn })
app.use(pinia)
app.use(router)
// Register echarts globally for HopeChart component
app.provide('echarts', echarts)
app.mount('#app')