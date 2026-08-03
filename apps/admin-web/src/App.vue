<template>
  <div v-if="authStore.user" style="height: 100vh;">
    <!-- Full app layout (only shown when logged in) -->
    <el-container style="height: 100%;">
      <!-- Sidebar -->
      <el-aside width="220px" class="sidebar">
        <div class="sidebar-top-bar"></div>
        <div class="sidebar-logo">
          <span class="logo-brand">Eregen</span>
          <span class="logo-cn">颐贞</span>
        </div>
        <el-menu :default-active="activeMenu" background-color="transparent" text-color="rgba(255,255,255,0.7)" active-text-color="#fff" router class="sidebar-menu">
          <el-divider class="menu-divider" content-position="left"><span class="section-label">概览</span></el-divider>
          <el-menu-item index="/dashboard">
            <el-icon><DataAnalysis /></el-icon><span>仪表盘</span>
          </el-menu-item>
          <el-divider class="menu-divider" content-position="left"><span class="section-label">设备管理</span></el-divider>
          <el-menu-item index="/devices">
            <el-icon><Watch /></el-icon><span>手环设备</span>
          </el-menu-item>
          <el-menu-item index="/pillboxes">
            <el-icon><PieChart /></el-icon><span>药盒设备</span>
          </el-menu-item>
          <el-menu-item index="/ota">
            <el-icon><Download /></el-icon><span>固件OTA</span>
          </el-menu-item>
          <el-divider class="menu-divider" content-position="left"><span class="section-label">用药管理</span></el-divider>
          <el-menu-item index="/medication">
            <el-icon><Document /></el-icon><span>用药规则</span>
          </el-menu-item>
          <el-divider class="menu-divider" content-position="left"><span class="section-label">用户管理</span></el-divider>
          <el-menu-item index="/users">
            <el-icon><User /></el-icon><span>家属用户</span>
          </el-menu-item>
          <el-menu-item index="/elderly">
            <el-icon><Avatar /></el-icon><span>老人档案</span>
          </el-menu-item>
          <el-menu-item index="/institutions">
            <el-icon><OfficeBuilding /></el-icon><span>机构管理</span>
          </el-menu-item>
          <el-divider class="menu-divider" content-position="left"><span class="section-label">医疗管理</span></el-divider>
          <el-menu-item index="/medical">
            <el-icon><FirstAidKit /></el-icon><span>医疗腕带</span>
          </el-menu-item>
          <el-menu-item index="/regulatory">
            <el-icon><Checked /></el-icon><span>监管看板</span>
          </el-menu-item>
          <el-menu-item index="/community-wb">
            <el-icon><Avatar /></el-icon><span>社区老人</span>
          </el-menu-item>
          <el-divider class="menu-divider" content-position="left"><span class="section-label">运营</span></el-divider>
          <el-menu-item index="/alerts">
            <el-icon><Bell /></el-icon><span>告警中心</span>
          </el-menu-item>
          <el-menu-item index="/subscriptions">
            <el-icon><List /></el-icon><span>订阅管理</span>
          </el-menu-item>
          <el-menu-item index="/analytics">
            <el-icon><TrendCharts /></el-icon><span>数据分析</span>
          </el-menu-item>
          <el-divider class="menu-divider" content-position="left"><span class="section-label">系统</span></el-divider>
          <el-menu-item index="/settings">
            <el-icon><Setting /></el-icon><span>系统设置</span>
          </el-menu-item>
        </el-menu>
        <div class="sidebar-footer">
          <el-avatar size="small" style="background: linear-gradient(135deg, #79A3D0, #165DFF);">管</el-avatar>
          <div>
            <div class="footer-name">{{ authStore.user.name }}</div>
            <div class="footer-role">{{ authStore.user.role === 'super_admin' ? '超级管理员' : '管理员' }}</div>
          </div>
        </div>
      </el-aside>

      <el-container>
        <!-- Top bar -->
        <el-header class="topbar">
          <div class="breadcrumb">
            <span class="breadcrumb-dot"></span>
            {{ currentBreadcrumb }}
          </div>
          <div class="topbar-right">
            <div class="topbar-icon" title="搜索">
              <el-icon :size="18"><Search /></el-icon>
            </div>
            <el-badge :value="3" :max="99">
              <div class="topbar-icon" title="通知">
                <el-icon :size="18"><Bell /></el-icon>
              </div>
            </el-badge>
            <div class="topbar-icon" title="主题">
              <el-icon :size="18"><Moon /></el-icon>
            </div>
            <el-button type="danger" @click="handleLogout" plain size="small" class="logout-btn">退出</el-button>
          </div>
        </el-header>

        <!-- Main content -->
        <el-main class="main-content">
          <router-view />
        </el-main>
      </el-container>
    </el-container>
  </div>
  <div v-else class="login-page-wrapper">
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
    if (route.path !== '/login') {
      router.push({ path: '/login', query: { redirect: route.path } })
    }
  }
})
</script>

<style scoped>
@import url('https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700;800&family=Noto+Sans+SC:wght@400;500;600;700&display=swap');

/* Global base */
html, body, #app {
  margin: 0;
  padding: 0;
  height: 100%;
  font-family: 'Inter', 'Noto Sans SC', -apple-system, BlinkMacSystemFont, 'PingFang SC', sans-serif;
  background-color: #FFFFFF;
}

/* ==================== SIDEBAR ==================== */
.sidebar {
  background: linear-gradient(180deg, #0F52BA 0%, #165DFF 50%, #1A4FD6 100%);
  display: flex;
  flex-direction: column;
  position: relative;
  box-shadow: 4px 0 20px rgba(0,0,0,0.12);
  z-index: 100;
}

.sidebar-top-bar {
  height: 3px;
  background: linear-gradient(90deg, #79A3D0, #36D399, #79A3D0);
  flex-shrink: 0;
}

.sidebar-logo {
  height: 64px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  border-bottom: 1px solid rgba(255,255,255,0.1);
  padding: 0 20px;
}
.logo-brand {
  font-size: 20px;
  font-weight: 800;
  letter-spacing: -0.02em;
  background: linear-gradient(135deg, #FFFFFF 0%, #B8D4F0 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}
.logo-cn {
  font-size: 13px;
  font-weight: 500;
  color: rgba(255,255,255,0.5);
  letter-spacing: 0.05em;
}

/* Menu items */
.sidebar-menu {
  flex: 1;
  overflow-y: auto;
  padding: 8px 0;
}
.sidebar-menu :deep(.el-menu) {
  border: none !important;
  background: transparent !important;
}
.sidebar-menu :deep(.el-menu-item) {
  height: 42px !important;
  line-height: 42px !important;
  margin: 2px 8px !important;
  border-radius: 8px !important;
  font-size: 13.5px !important;
  color: rgba(255,255,255,0.7) !important;
  transition: all 0.2s ease !important;
}
.sidebar-menu :deep(.el-menu-item:hover) {
  background: rgba(255,255,255,0.1) !important;
  color: #FFFFFF !important;
  transform: translateX(2px);
}
.sidebar-menu :deep(.el-menu-item.is-active) {
  background: rgba(255,255,255,0.15) !important;
  color: #FFFFFF !important;
  font-weight: 600;
}
.sidebar-menu :deep(.el-menu-item.is-active)::before {
  content: '';
  position: absolute;
  left: 0;
  top: 50%;
  transform: translateY(-50%);
  width: 3px;
  height: 20px;
  background: #FFFFFF;
  border-radius: 0 3px 3px 0;
}

/* Section dividers */
.menu-divider {
  margin: 8px 16px 4px !important;
  border-color: rgba(255,255,255,0.08) !important;
}
.section-label {
  font-size: 10px !important;
  color: rgba(255,255,255,0.35) !important;
  letter-spacing: 0.1em !important;
  text-transform: uppercase;
  font-weight: 600;
}

/* Footer user info */
.sidebar-footer {
  padding: 16px 20px;
  border-top: 1px solid rgba(255,255,255,0.08);
  display: flex;
  align-items: center;
  gap: 12px;
  background: rgba(0,0,0,0.1);
  margin: 0 0;
}
.footer-name {
  font-size: 12px;
  font-weight: 600;
  color: rgba(255,255,255,0.9);
}
.footer-role {
  font-size: 11px;
  color: rgba(255,255,255,0.4);
  margin-top: 1px;
}

/* ==================== TOPBAR ==================== */
.topbar {
  height: 64px;
  background: rgba(255,255,255,0.95);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  border-bottom: 1px solid #E8ECF1;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 32px;
  position: sticky;
  top: 0;
  z-index: 50;
  box-shadow: 0 1px 3px rgba(0,0,0,0.04);
}
.breadcrumb {
  font-size: 13px;
  color: #64748B;
  font-weight: 500;
  display: flex;
  align-items: center;
  gap: 8px;
}
.breadcrumb-dot {
  width: 6px;
  height: 6px;
  min-width: 6px;
  border-radius: 50%;
  background: linear-gradient(135deg, #165DFF, #79A3D0);
}
.breadcrumb span:last-child {
  color: #0F172A;
  font-weight: 600;
}
.topbar-right {
  display: flex;
  align-items: center;
  gap: 8px;
}
.topbar-icon {
  cursor: pointer;
  color: #64748B;
  transition: all 0.2s ease;
  padding: 8px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
}
.topbar-icon:hover {
  color: #165DFF;
  background: #EFF6FF;
}
.logout-btn {
  margin-left: 8px;
  border-radius: 8px !important;
  font-size: 13px;
  padding: 6px 16px !important;
}

/* Badge */
.el-badge__content {
  background: #F87272 !important;
  font-size: 10px !important;
  padding: 1px 5px !important;
  min-width: 0 !important;
  height: auto !important;
  border-radius: 10px !important;
  border: 1px solid #FFFFFF;
}

/* ==================== MAIN CONTENT ==================== */
.main-content {
  background: #FFFFFF;
  padding: 28px;
  overflow-y: auto;
  height: calc(100vh - 64px);
}

/* ==================== LOGIN PAGE WRAPPER ==================== */
.login-page-wrapper {
  min-height: 100vh;
}

/* ==================== RESPONSIVE ==================== */
@media (max-width: 768px) {
  .sidebar { width: 70px !important; }
  .sidebar-logo .logo-brand,
  .sidebar-logo .logo-cn,
  .sidebar .el-menu-item span,
  .section-label { display: none; }
  .sidebar-menu :deep(.el-menu-item) {
    justify-content: center !important;
    padding: 0 !important;
    margin: 2px 6px !important;
  }
  .sidebar-footer { justify-content: center; }
  .main-content { padding: 16px; }
  .topbar { padding: 0 16px; }
}
</style>
