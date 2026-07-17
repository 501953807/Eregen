<template>
  <el-container style="height: 100vh;">
    <!-- Sidebar -->
    <el-aside width="220px" class="sidebar">
      <div class="sidebar-logo"><span>Eregen</span> 颐贞</div>
      <el-menu :default-active="activeMenu" background-color="#001529" text-color="rgba(255,255,255,0.65)" active-text-color="#fff" router>
        <el-divider content-position="left" style="color:rgba(255,255,255,0.1);"><span style="font-size:10px;color:rgba(255,255,255,0.35);letter-spacing:1px;">概览</span></el-divider>
        <el-menu-item index="/dashboard">
          <el-icon><DataAnalysis /></el-icon><span>仪表盘</span>
        </el-menu-item>
        <el-divider content-position="left" style="color:rgba(255,255,255,0.1);"><span style="font-size:10px;color:rgba(255,255,255,0.35);letter-spacing:1px;">设备管理</span></el-divider>
        <el-menu-item index="/devices">
          <el-icon><Watch /></el-icon><span>手环设备</span>
        </el-menu-item>
        <el-menu-item index="/pillboxes">
          <el-icon><PieChart /></el-icon><span>药盒设备</span>
        </el-menu-item>
        <el-menu-item index="/ota">
          <el-icon><Download /></el-icon><span>固件OTA</span>
        </el-menu-item>
        <el-divider content-position="left" style="color:rgba(255,255,255,0.1);"><span style="font-size:10px;color:rgba(255,255,255,0.35);letter-spacing:1px;">用户管理</span></el-divider>
        <el-menu-item index="/users">
          <el-icon><User /></el-icon><span>家属用户</span>
        </el-menu-item>
        <el-menu-item index="/elderly">
          <el-icon><Avatar /></el-icon><span>老人档案</span>
        </el-menu-item>
        <el-menu-item index="/institutions">
          <el-icon><OfficeBuilding /></el-icon><span>机构管理</span>
        </el-menu-item>
        <el-divider content-position="left" style="color:rgba(255,255,255,0.1);"><span style="font-size:10px;color:rgba(255,255,255,0.35);letter-spacing:1px;">运营</span></el-divider>
        <el-menu-item index="/alerts">
          <el-icon><Bell /></el-icon><span>告警中心</span>
        </el-menu-item>
        <el-menu-item index="/subscriptions">
          <el-icon><List /></el-icon><span>订阅管理</span>
        </el-menu-item>
        <el-menu-item index="/analytics">
          <el-icon><TrendCharts /></el-icon><span>数据分析</span>
        </el-menu-item>
        <el-divider content-position="left" style="color:rgba(255,255,255,0.1);"><span style="font-size:10px;color:rgba(255,255,255,0.35);letter-spacing:1px;">系统</span></el-divider>
        <el-menu-item index="/settings">
          <el-icon><Setting /></el-icon><span>系统设置</span>
        </el-menu-item>
      </el-menu>
      <div class="sidebar-footer">
        <el-avatar size="small" style="background:#4A90D9;">管</el-avatar>
        <div>
          <div style="font-size:12px;font-weight:600;">管理员</div>
          <div style="font-size:11px;color:rgba(255,255,255,0.4);">超级管理员</div>
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
        </div>
      </el-header>

      <!-- Main content -->
      <el-main class="main-content">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import {
  DataAnalysis, Watch, PieChart, Download, User, Avatar,
  OfficeBuilding, Bell, List, TrendCharts, Setting, Search, Moon
} from '@element-plus/icons-vue'

const route = useRoute()
const activeMenu = computed(() => route.path)
const currentBreadcrumb = computed(() => {
  const map: Record<string, string> = {
    '/dashboard': '首页 / 仪表盘总览',
    '/devices': '设备管理 / 手环设备',
    '/pillboxes': '设备管理 / 药盒设备',
    '/subscriptions': '运营管理 / 订阅管理',
    '/users': '用户管理 / 全部用户',
    '/institutions': '用户管理 / 机构管理',
    '/alerts': '告警中心 / 告警列表',
    '/analytics': '数据分析 / 概览',
    '/settings': '系统设置 / 配置',
    '/ota': '设备管理 / 固件OTA',
    '/elderly': '用户管理 / 老人档案',
  }
  return map[route.path] || route.path
})
</script>

<style>
html, body, #app { margin: 0; padding: 0; height: 100%; font-family: -apple-system, 'PingFang SC', sans-serif; }
.sidebar { background: #001529; display: flex; flex-direction: column; }
.sidebar-logo { height: 64px; display: flex; align-items: center; justify-content: center; font-size: 20px; font-weight: 700; letter-spacing: 2px; border-bottom: 1px solid rgba(255,255,255,0.08); }
.sidebar-logo span { color: #4A90D9; }
.sidebar .el-menu { border-right: none; }
.sidebar-footer { margin-top: auto; padding: 16px 24px; border-top: 1px solid rgba(255,255,255,0.08); display: flex; align-items: center; gap: 10px; }
.topbar { height: 64px; background: #fff; border-bottom: 1px solid #e8e8e8; display: flex; align-items: center; justify-content: space-between; padding: 0 24px; position: sticky; top: 0; z-index: 5; }
.breadcrumb { font-size: 14px; color: #999; }
.breadcrumb span { color: #333; font-weight: 600; }
.topbar-right { display: flex; align-items: center; gap: 20px; }
.main-content { background: #f0f2f5; padding: 24px; }
.el-divider__text { font-size: 10px; }
</style>
