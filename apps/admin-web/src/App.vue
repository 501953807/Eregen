<template>
  <div v-if="authStore.user" style="height: 100vh;">
    <!-- Full app layout (only shown when logged in) -->
    <el-container style="height: 100%;">
      <!-- Sidebar -->
      <el-aside width="220px" class="sidebar">
        <div class="sidebar-logo"><span>Eregen</span> 颐贞</div>
        <el-menu :default-active="activeMenu" background-color="#1F2937" text-color="rgba(255,255,255,0.8)" active-text-color="#fff" router>
          <el-divider content-position="left" style="color:rgba(255,255,255,0.1);"><span style="font-size:10px;color:rgba(255,255,255,0.4);letter-spacing:1px;">概览</span></el-divider>
          <el-menu-item index="/dashboard">
            <el-icon><DataAnalysis /></el-icon><span>仪表盘</span>
          </el-menu-item>
          <el-divider content-position="left" style="color:rgba(255,255,255,0.1);"><span style="font-size:10px;color:rgba(255,255,255,0.4);letter-spacing:1px;">设备管理</span></el-divider>
          <el-menu-item index="/devices">
            <el-icon><Watch /></el-icon><span>手环设备</span>
          </el-menu-item>
          <el-menu-item index="/pillboxes">
            <el-icon><PieChart /></el-icon><span>药盒设备</span>
          </el-menu-item>
          <el-menu-item index="/ota">
            <el-icon><Download /></el-icon><span>固件OTA</span>
          </el-menu-item>
          <el-divider content-position="left" style="color:rgba(255,255,255,0.1);"><span style="font-size:10px;color:rgba(255,255,255,0.4);letter-spacing:1px;">用药管理</span></el-divider>
          <el-menu-item index="/medication">
            <el-icon><Document /></el-icon><span>用药规则</span>
          </el-menu-item>
          <el-divider content-position="left" style="color:rgba(255,255,255,0.1);"><span style="font-size:10px;color:rgba(255,255,255,0.4);letter-spacing:1px;">用户管理</span></el-divider>
          <el-menu-item index="/users">
            <el-icon><User /></el-icon><span>家属用户</span>
          </el-menu-item>
          <el-menu-item index="/elderly">
            <el-icon><Avatar /></el-icon><span>老人档案</span>
          </el-menu-item>
          <el-menu-item index="/institutions">
            <el-icon><OfficeBuilding /></el-icon><span>机构管理</span>
          </el-menu-item>
          <el-divider content-position="left" style="color:rgba(255,255,255,0.1);"><span style="font-size:10px;color:rgba(255,255,255,0.4);letter-spacing:1px;">医疗管理</span></el-divider>
          <el-menu-item index="/medical">
            <el-icon><FirstAidKit /></el-icon><span>医疗腕带</span>
          </el-menu-item>
          <el-menu-item index="/regulatory">
            <el-icon><Checked /></el-icon><span>监管看板</span>
          </el-menu-item>
          <el-menu-item index="/community-wb">
            <el-icon><Avatar /></el-icon><span>社区老人</span>
          </el-menu-item>
          <el-divider content-position="left" style="color:rgba(255,255,255,0.1);"><span style="font-size:10px;color:rgba(255,255,255,0.4);letter-spacing:1px;">运营</span></el-divider>
          <el-menu-item index="/alerts">
            <el-icon><Bell /></el-icon><span>告警中心</span>
          </el-menu-item>
          <el-menu-item index="/subscriptions">
            <el-icon><List /></el-icon><span>订阅管理</span>
          </el-menu-item>
          <el-menu-item index="/analytics">
            <el-icon><TrendCharts /></el-icon><span>数据分析</span>
          </el-menu-item>
          <el-divider content-position="left" style="color:rgba(255,255,255,0.1);"><span style="font-size:10px;color:rgba(255,255,255,0.4);letter-spacing:1px;">系统</span></el-divider>
          <el-menu-item index="/settings">
            <el-icon><Setting /></el-icon><span>系统设置</span>
          </el-menu-item>
        </el-menu>
        <div class="sidebar-footer">
          <el-avatar size="small" style="background:#F59E0B;">管</el-avatar>
          <div>
            <div style="font-size:12px;font-weight:600;">{{ authStore.user.name }}</div>
            <div style="font-size:11px;color:rgba(255,255,255,0.4);">{{ authStore.user.role === 'super_admin' ? '超级管理员' : '管理员' }}</div>
          </div>
        </div>
      </el-aside>

      <el-container>
        <!-- Top bar -->
        <el-header class="topbar">
          <div class="breadcrumb">{{ currentBreadcrumb }}</div>
          <div class="topbar-right">
            <el-icon :size="18" style="cursor:pointer;">Search</el-icon>
            <el-badge :value="3" :max="99">
              <el-icon :size="18" style="cursor:pointer;">Bell</el-icon>
            </el-badge>
            <el-icon :size="18" style="cursor:pointer;">Moon</el-icon>
            <el-button type="danger" @click="handleLogout" plain>退出</el-button>
          </div>
        </el-header>

        <!-- Main content -->
        <el-main class="main-content">
          <router-view />
        </el-main>
      </el-container>
    </el-container>
  </div>
  <div v-else class="login-page">
    <LoginView />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  DataAnalysis, Watch, PieChart, Download, User, Avatar,
  OfficeBuilding, Bell, List, TrendCharts, Setting, Search, Moon,
  Document, FirstAidKit, Checked
} from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import LoginView from '@/views/Login.vue'
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()
const route = useRoute()
const router = useRouter()

const activeMenu = computed(() => route.path)

const currentBreadcrumb = computed(() => {
  const map: Record<string, string> = {
    '/dashboard': '首页 / 仪表盘总览',
    '/devices': '设备管理 / 手环设备',
    '/pillboxes': '设备管理 / 药盒设备',
    '/ota': '设备管理 / 固件OTA',
    '/medication': '用药管理 / 用药规则',
    '/subscriptions': '运营管理 / 订阅管理',
    '/users': '用户管理 / 全部用户',
    '/institutions': '用户管理 / 机构管理',
    '/elderly': '用户管理 / 老人档案',
    '/alerts': '告警中心 / 告警列表',
    '/analytics': '数据分析 / 概览',
    '/settings': '系统设置 / 配置',
    '/medical': '医疗管理 / 医疗腕带',
    '/regulatory': '医疗管理 / 监管看板',
    '/community-wb': '医疗管理 / 社区老人',
  }
  return map[route.path] || route.path
})

async function handleLogout() {
  await authStore.logout()
  ElMessage.info('已安全退出')
}

// Check authentication on mount
onMounted(async () => {
  if (!authStore.isLoggedIn()) {
    // If not logged in and trying to access a protected route, redirect to login
    if (route.path !== '/login') {
      router.push({ path: '/login', query: { redirect: route.path } })
    }
  }
})
</script>

<style scoped>
@import url('https://fonts.googleapis.com/css2?family=Noto+Sans+SC:wght@300;400;500;600;700&display=swap');

/* Global base styles */
html, body, #app {
  margin: 0;
  padding: 0;
  height: 100%;
  font-family: 'Noto Sans SC', -apple-system, BlinkMacSystemFont, 'PingFang SC', sans-serif;
  background-color: var(--el-bg-color);
}

/* Upgrade sidebar with warmer palette */
.sidebar {
  background: linear-gradient(180deg, #1F2937 0%, #111827 100%);
  display: flex;
  flex-direction: column;
  box-shadow: 2px 0 8px rgba(0,0,0,0.1);
}

.sidebar-logo {
  height: 64px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
  font-weight: 700;
  letter-spacing: -0.02em;
  border-bottom: 1px solid rgba(255,255,255,0.08);
  background: linear-gradient(to bottom, rgba(31,41,55,1), rgba(17,24,39,1));
}
.sidebar-logo span { color: #F59E0B; }

/* Sidebar menu item hover enhancement */
.el-menu-item {
  position: relative;
  transition: all 0.2s ease;
}
.el-menu-item:hover {
  background-color: rgba(245, 158, 11, 0.12);
  transform: translateX(4px);
}

/* Footer area styling */
.sidebar-footer {
  margin-top: auto;
  padding: 16px 20px;
  border-top: 1px solid rgba(255,255,255,0.08);
  display: flex;
  align-items: center;
  gap: 12px;
  background: rgba(0,0,0,0.1);
}

/* Top bar with shadow */
.topbar {
  height: 64px;
  background: white;
  border-bottom: 1px solid var(--el-border-color-light);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 32px;
  position: sticky; top: 0; z-index: 50;
  box-shadow: 0 1px 3px rgba(0,0,0,0.05);
  backdrop-filter: blur(8px);
  -webkit-backdrop-filter: blur(8px);
}
.breadcrumb {
  font-size: 14px;
  color: var(--el-text-color-secondary);
  font-weight: 500;
}
.breadcrumb span {
  color: var(--el-text-color-primary);
  font-weight: 600;
  margin-left: 8px;
}
.topbar-right {
  display: flex;
  align-items: center;
  gap: 20px;
}
.topbar-icon {
  cursor: pointer;
  color: var(--el-text-color-secondary);
  transition: color 0.2s ease;
  padding: 8px;
  border-radius: 8px;
}
.topbar-icon:hover {
  color: var(--el-color-primary);
  background-color: var(--el-color-primary-light);
}

/* Badge for notifications */
.el-badge__content {
  background: var(--el-color-danger);
  font-size: 11px !important;
  padding: 2px 6px !important;
  min-width: 0 !important;
  height: auto !important;
  border-radius: 10px;
}

/* Main content area with improved spacing */
.main-content {
  background: var(--el-bg-color);
  padding: 32px;
  overflow-y: auto;
  height: calc(100vh - 64px);
}

.login-page {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  background: linear-gradient(135deg, #FEF3C7 0%, #FDE68A 100%);
}

/* Enhanced button styles */.el-button--primary {
  --el-button-primary-hover: var(--el-color-primary-dark);
  --el-button-primary-active: var(--el-color-primary-light);
  --el-button-focus-transform: translateY(-2px);
  --el-button-active-transform: translateY(0) scale(0.98);
  transition: all var(--el-transition-duration) var(--el-transition-easing);
}

/* Card hover effect enhancement */
.el-card:hover {
  box-shadow: var(--el-shadow-lg);
  transform: translateY(-2px);
  transition: all 0.3s ease;
}

/* Focus ring for accessibility */
*:focus-visible {
  outline: 2px solid var(--el-color-primary);
  outline-offset: 2px;
}

/* Responsive adjustments */
@media (max-width: 768px) {
  .sidebar { width: 70px; }
  .sidebar-logo span { display: none; }
  .el-menu-item span { display: none; }
  .sidebar .el-menu { justify-content: center; }
  .main-content { padding: 16px; }
  .topbar { padding: 0 16px; }
  .breadcrumb { display: none; }
}
</style>