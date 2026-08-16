<template>
  <header class="hope-topbar">
    <div class="topbar-left">
      <div class="search-box">
        <svg class="search-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <circle cx="11" cy="11" r="8"/>
          <path d="M21 21l-4.35-4.35"/>
        </svg>
        <input type="text" placeholder="Search..." class="search-input" />
      </div>
    </div>

    <div class="topbar-right">
      <button class="topbar-btn" title="Notifications">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M18 8A6 6 0 006 8c0 7-3 9-3 9h18s-3-2-3-9"/>
          <path d="M13.73 21a2 2 0 01-3.46 0"/>
        </svg>
        <span class="badge">3</span>
      </button>

      <button class="topbar-btn" title="Messages">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M21 15a2 2 0 01-2 2H7l-4 4V5a2 2 0 012-2h14a2 2 0 012 2z"/>
        </svg>
      </button>

      <el-dropdown trigger="click" @command="handleCommand">
        <div class="user-menu">
          <div class="user-avatar-small">
            <span>{{ authStore.user?.name?.charAt(0) || '管' }}</span>
          </div>
          <span class="user-name">{{ authStore.user?.name }}</span>
          <svg class="dropdown-arrow" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <polyline points="6 9 12 15 18 9"/>
          </svg>
        </div>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item command="profile">个人资料</el-dropdown-item>
            <el-dropdown-item command="settings">系统设置</el-dropdown-item>
            <el-dropdown-item divided command="logout">退出登录</el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
    </div>
  </header>
</template>

<script setup lang="ts">
import { ElMessage } from 'element-plus'
import { useAuthStore } from '@/stores/auth'
import { useRouter } from 'vue-router'

const authStore = useAuthStore()
const router = useRouter()

function handleCommand(command: string) {
  if (command === 'logout') {
    authStore.logout()
    ElMessage.info('已安全退出')
    router.push('/login')
  } else if (command === 'settings') {
    router.push('/settings')
  }
}
</script>

<style scoped>
.hope-topbar {
  height: 64px;
  background: var(--hope-surface);
  border-bottom: 1px solid var(--hope-border);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 1.5rem;
  position: sticky;
  top: 0;
  z-index: 90;
  box-shadow: var(--hope-shadow-sm);
}

.topbar-left {
  display: flex;
  align-items: center;
}

.search-box {
  position: relative;
  display: flex;
  align-items: center;
}

.search-icon {
  position: absolute;
  left: 12px;
  width: 16px;
  height: 16px;
  color: var(--hope-text-muted);
}

.search-input {
  width: 200px;
  height: 36px;
  padding: 0 12px 0 36px;
  border: 1px solid var(--hope-border);
  border-radius: var(--hope-radius-sm);
  background: var(--hope-surface-light);
  font-size: 0.875rem;
  color: var(--hope-text);
  transition: all 0.2s ease;
  outline: none;
}

.search-input:focus {
  border-color: var(--hope-primary);
  box-shadow: var(--hope-shadow-input-focus);
  width: 240px;
}

.search-input::placeholder {
  color: var(--hope-text-muted);
}

.topbar-right {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.topbar-btn {
  width: 36px;
  height: 36px;
  border-radius: var(--hope-radius-sm);
  border: 1px solid var(--hope-border);
  background: var(--hope-surface);
  color: var(--hope-text-secondary);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.15s ease;
  position: relative;
}

.topbar-btn:hover {
  background: var(--hope-primary-lighter);
  color: var(--hope-primary);
  border-color: var(--hope-primary-light);
}

.badge {
  position: absolute;
  top: -4px;
  right: -4px;
  width: 18px;
  height: 18px;
  border-radius: 50%;
  background: var(--hope-danger);
  color: white;
  font-size: 0.625rem;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 2px solid var(--hope-surface);
}

.user-menu {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 4px 8px;
  border-radius: var(--hope-radius-md);
  cursor: pointer;
  transition: background 0.15s ease;
  border: 1px solid var(--hope-border);
}

.user-menu:hover {
  background: var(--hope-primary-lighter);
  border-color: var(--hope-primary-light);
}

.user-avatar-small {
  width: 28px;
  height: 28px;
  border-radius: 50%;
  background: var(--hope-primary-gradient);
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  font-size: 0.75rem;
  font-weight: 700;
}

.user-name {
  font-size: 0.8125rem;
  font-weight: 500;
  color: var(--hope-text);
  max-width: 80px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.dropdown-arrow {
  width: 12px;
  height: 12px;
  color: var(--hope-text-muted);
}
</style>
