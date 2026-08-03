<template>
  <div class="login-page">
    <!-- Left decorative panel -->
    <div class="login-brand">
      <div class="brand-bg-shapes">
        <div class="shape shape-1"></div>
        <div class="shape shape-2"></div>
        <div class="shape shape-3"></div>
        <div class="shape shape-4"></div>
      </div>
      <div class="brand-content">
        <div class="brand-logo">
          <div class="logo-icon">
            <svg width="32" height="32" viewBox="0 0 32 32" fill="none">
              <rect width="32" height="32" rx="8" fill="rgba(255,255,255,0.15)"/>
              <path d="M16 8L10 14v6l6 3 6-3v-6L16 8z" fill="rgba(255,255,255,0.9)" stroke="rgba(255,255,255,0.5)" stroke-width="0.5"/>
              <circle cx="16" cy="17" r="2" fill="rgba(22,93,255,0.8)"/>
            </svg>
          </div>
          <div>
            <div class="logo-text">Eregen</div>
            <div class="logo-sub">颐贞 · 智能健康</div>
          </div>
        </div>
        <div class="brand-features">
          <div class="feature-item">
            <div class="feature-dot"></div>
            <div>
              <div class="feature-title">实时健康监控</div>
              <div class="feature-desc">手环/药盒设备 24h 数据采集</div>
            </div>
          </div>
          <div class="feature-item">
            <div class="feature-dot"></div>
            <div>
              <div class="feature-title">智能告警中心</div>
              <div class="feature-desc">跌倒 / 心率异常 / SOS 即时推送</div>
            </div>
          </div>
          <div class="feature-item">
            <div class="feature-dot"></div>
            <div>
              <div class="feature-title">用药依从管理</div>
              <div class="feature-desc">规则配置 + 漏服提醒 + 数据统计</div>
            </div>
          </div>
          <div class="feature-item">
            <div class="feature-dot"></div>
            <div>
              <div class="feature-title">机构协同运营</div>
              <div class="feature-desc">多机构数据汇总 · 监管合规看板</div>
            </div>
          </div>
        </div>
        <div class="brand-footer">
          <span>© 2024 Eregen · 颐贞智能健康平台</span>
        </div>
      </div>
    </div>

    <!-- Right login form panel -->
    <div class="login-form-panel">
      <div class="form-card">
        <div class="form-left-accent"></div>
        <div class="form-inner">
          <div class="login-header">
            <h1 class="login-title">欢迎登录</h1>
            <p class="login-subtitle">颐贞智能健康 · 管理后台</p>
          </div>

          <div class="login-tabs">
            <div
              class="tab-item"
              :class="{ active: activeTab === 'email' }"
              @click="activeTab = 'email'"
            >账号登录</div>
            <div
              class="tab-item"
              :class="{ active: activeTab === 'phone' }"
              @click="activeTab = 'phone'"
            >手机验证码</div>
          </div>

          <!-- Email / Password Form -->
          <el-form
            v-show="activeTab === 'email'"
            ref="emailFormEl"
            :model="emailForm"
            :rules="emailRules"
            label-width="0"
            size="large"
            class="login-form"
          >
            <el-form-item prop="email">
              <el-input
                v-model="emailForm.email"
                placeholder="请输入邮箱"
                :prefix-icon="User"
                clearable
              />
            </el-form-item>
            <el-form-item prop="password">
              <el-input
                v-model="emailForm.password"
                type="password"
                placeholder="请输入密码"
                :prefix-icon="Lock"
                show-password
                @keyup.enter="submitEmailLogin"
              />
            </el-form-item>
            <el-form-item>
              <el-button
                type="primary"
                @click="submitEmailLogin"
                :loading="authStore.loading"
                class="login-btn"
              >{{ authStore.loading ? '登录中...' : '登 录' }}</el-button>
            </el-form-item>
            <div v-if="authStore.error" class="error-message">{{ authStore.error }}</div>
          </el-form>

          <!-- Phone / OTP Form -->
          <el-form
            v-show="activeTab === 'phone'"
            ref="phoneFormEl"
            :model="phoneForm"
            :rules="phoneRules"
            label-width="0"
            size="large"
            class="login-form"
          >
            <el-form-item prop="phone">
              <el-input
                v-model="phoneForm.phone"
                placeholder="请输入手机号"
                :prefix-icon="Phone"
                type="tel"
                clearable
              />
            </el-form-item>
            <el-form-item prop="otp">
              <el-input
                v-model="phoneForm.otp"
                placeholder="6位验证码"
                :prefix-icon="Message"
                maxlength="6"
                type="digit"
                class="otp-input"
              />
              <el-button
                type="primary"
                plain
                @click="sendOtp"
                :disabled="countdown > 0 || authStore.loading"
                class="otp-btn"
              >{{ countdown > 0 ? countdown + 's' : '获取验证码' }}</el-button>
            </el-form-item>
            <el-form-item>
              <el-button
                type="primary"
                @click="submitPhoneLogin"
                :loading="authStore.loading"
                class="login-btn"
              >{{ authStore.loading ? '登录中...' : '登 录' }}</el-button>
            </el-form-item>
            <div v-if="authStore.error" class="error-message">{{ authStore.error }}</div>
          </el-form>

          <div class="test-accounts">
            <div class="test-account-card">
              <div class="test-label">测试账号（邮箱）</div>
              <div class="test-cred">admin@eregen.com / Admin@123</div>
            </div>
            <div class="test-account-card">
              <div class="test-label">测试账号（手机）</div>
              <div class="test-cred">13800000002 / 123456</div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import type { FormRules } from 'element-plus'
import { ElMessage } from 'element-plus'
import { User, Lock, Phone, Message } from '@element-plus/icons-vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()
const route = useRoute()
const router = useRouter()

const activeTab = ref<'email' | 'phone'>('email')

// --- Email / Password ---
const emailFormEl = ref<any>(null)
const emailForm = ref({ email: '', password: '' })
const emailRules = computed<FormRules>(() => ({
  email: [
    { required: true, message: '请输入邮箱', trigger: 'blur' },
    { pattern: /^[^\s@]+@[^\s@]+\.[^\s@]+$/, message: '邮箱格式不正确', trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, message: '密码至少6位', trigger: 'blur' }
  ]
}))

async function submitEmailLogin() {
  if (!emailFormEl.value) return
  const valid = await emailFormEl.value.validate()
  if (!valid) return
  try {
    await authStore.login({
      method: 'email',
      credential: emailForm.value.email,
      secret: emailForm.value.password
    })
    const to = route.query.redirect || '/dashboard'
    router.push(to as string)
  } catch (err: any) {
    // error already set by store
  }
}

// --- Phone / OTP ---
const phoneFormEl = ref<any>(null)
const phoneForm = ref({ phone: '', otp: '' })
const phoneRules = computed<FormRules>(() => ({
  phone: [
    { required: true, message: '请输入手机号', trigger: 'blur' },
    { pattern: /^(\+86|86)?1[3-9]\d{9}$/, message: '请输入正确的中国大陆手机号', trigger: 'blur' }
  ],
  otp: [
    { required: true, message: '请输入验证码', trigger: 'blur' },
    { pattern: /^\d{6}$/, message: '验证码应为6位数字', trigger: 'blur' }
  ]
}))

const countdown = ref(0)
const timerRef = ref<number | null>(null)

function sendOtp() {
  if (!phoneForm.value.phone) {
    ElMessage.warning('请先输入手机号')
    return
  }
  ElMessage.success('验证码已发送，测试用：123456')
  countdown.value = 60
  timerRef.value = window.setInterval(() => {
    countdown.value--
    if (countdown.value <= 0) {
      if (timerRef.value) clearInterval(timerRef.value)
      timerRef.value = null
    }
  }, 1000)
}

async function submitPhoneLogin() {
  if (!phoneFormEl.value) return
  const valid = await phoneFormEl.value.validate()
  if (!valid) return
  try {
    await authStore.login({
      method: 'phone',
      credential: phoneForm.value.phone,
      secret: phoneForm.value.otp
    })
    const to = route.query.redirect || '/dashboard'
    router.push(to as string)
  } catch (err: any) {
    // error already set by store
  }
}

onMounted(() => {
  if (authStore.isLoggedIn()) {
    const to = route.query.redirect || '/dashboard'
    router.push(to as string)
  }
})

onBeforeUnmount(() => {
  if (timerRef.value) clearInterval(timerRef.value)
})
</script>

<style scoped>
.login-page {
  display: flex;
  min-height: 100vh;
  background: #F0F4F8;
  overflow: hidden;
}

/* ==================== LEFT DECORATIVE PANEL ==================== */
.login-brand {
  width: 440px;
  min-height: 100vh;
  background: linear-gradient(160deg, #0F52BA 0%, #165DFF 40%, #2B6DE8 100%);
  position: relative;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  padding: 48px 40px;
  flex-shrink: 0;
}

/* Floating decorative shapes */
.brand-bg-shapes {
  position: absolute;
  inset: 0;
  overflow: hidden;
  pointer-events: none;
}
.shape {
  position: absolute;
  border-radius: 50%;
  opacity: 0.08;
  background: #FFFFFF;
}
.shape-1 {
  width: 300px; height: 300px;
  top: -60px; right: -60px;
  animation: float 8s ease-in-out infinite;
}
.shape-2 {
  width: 180px; height: 180px;
  bottom: 120px; left: -40px;
  animation: float 10s ease-in-out infinite reverse;
}
.shape-3 {
  width: 100px; height: 100px;
  top: 45%; right: 30px;
  animation: float 6s ease-in-out infinite;
}
.shape-4 {
  width: 60px; height: 60px;
  bottom: 200px; right: 80px;
  opacity: 0.12;
  animation: float 7s ease-in-out infinite reverse;
}
@keyframes float {
  0%, 100% { transform: translateY(0); }
  50% { transform: translateY(-12px); }
}

.brand-content {
  position: relative;
  z-index: 1;
  display: flex;
  flex-direction: column;
  height: 100%;
}

/* Logo */
.brand-logo {
  display: flex;
  align-items: center;
  gap: 14px;
  margin-bottom: 48px;
}
.logo-icon {
  width: 44px;
  height: 44px;
  display: flex;
  align-items: center;
  justify-content: center;
}
.logo-text {
  font-size: 24px;
  font-weight: 800;
  color: #FFFFFF;
  letter-spacing: -0.02em;
  line-height: 1.1;
}
.logo-sub {
  font-size: 12px;
  color: rgba(255,255,255,0.6);
  letter-spacing: 0.08em;
  margin-top: 2px;
}

/* Feature list */
.brand-features {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 20px;
  margin-bottom: 40px;
}
.feature-item {
  display: flex;
  align-items: flex-start;
  gap: 14px;
}
.feature-dot {
  width: 8px;
  height: 8px;
  min-width: 8px;
  border-radius: 50%;
  background: rgba(255,255,255,0.5);
  margin-top: 6px;
}
.feature-title {
  font-size: 14px;
  font-weight: 600;
  color: #FFFFFF;
  margin-bottom: 2px;
}
.feature-desc {
  font-size: 12px;
  color: rgba(255,255,255,0.6);
  line-height: 1.5;
}

.brand-footer {
  font-size: 11px;
  color: rgba(255,255,255,0.35);
  letter-spacing: 0.02em;
}

/* ==================== RIGHT FORM PANEL ==================== */
.login-form-panel {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 40px;
  background: #FFFFFF;
}

.form-card {
  width: 100%;
  max-width: 420px;
  position: relative;
}
.form-left-accent {
  position: absolute;
  left: 0;
  top: 0;
  bottom: 0;
  width: 4px;
  border-radius: 0 2px 2px 0;
  background: linear-gradient(180deg, #165DFF 0%, #79A3D0 60%, #36D399 100%);
}

.form-inner {
  padding: 40px 36px;
}

/* Header */
.login-header {
  margin-bottom: 32px;
}
.login-title {
  font-size: 26px;
  font-weight: 700;
  color: #0F172A;
  margin: 0 0 6px;
  letter-spacing: -0.02em;
}
.login-subtitle {
  font-size: 13px;
  color: #64748B;
  margin: 0;
}

/* Tabs */
.login-tabs {
  display: flex;
  gap: 0;
  margin-bottom: 28px;
  border-bottom: 2px solid #F1F5F9;
}
.tab-item {
  flex: 1;
  text-align: center;
  padding: 12px 0;
  font-size: 14px;
  color: #94A3B8;
  cursor: pointer;
  transition: all 0.25s ease;
  border-bottom: 2px solid transparent;
  margin-bottom: -2px;
  font-weight: 500;
}
.tab-item:hover {
  color: #165DFF;
}
.tab-item.active {
  color: #165DFF;
  border-bottom-color: #165DFF;
  font-weight: 600;
}

/* Form */
.login-form {
  margin-bottom: 8px;
}
.login-form :deep(.el-form-item) {
  margin-bottom: 20px;
}
.login-form :deep(.el-input__wrapper) {
  padding: 0 12px;
  height: 44px;
}
.login-form :deep(.el-input__inner) {
  font-size: 14px;
}
.login-form :deep(.el-input__prefix) {
  margin-right: 4px;
  color: #94A3B8;
  transition: color 0.2s ease;
}
.login-form :deep(.el-input.is-focus .el-input__prefix) {
  color: #165DFF;
}

.login-btn {
  width: 100%;
  height: 46px;
  font-size: 15px;
  font-weight: 600;
  letter-spacing: 0.06em;
  border-radius: 8px !important;
}

/* OTP input row */
.otp-input {
  flex: 1;
  margin-right: 8px;
}
.otp-input :deep(.el-input__wrapper) {
  height: 44px;
}
.otp-btn {
  height: 44px;
  white-space: nowrap;
  border-radius: 8px !important;
}

/* Error */
.error-message {
  color: #DC2626;
  margin-top: 10px;
  font-size: 13px;
  text-align: center;
  padding: 8px 12px;
  background: #FEF2F2;
  border-radius: 8px;
  border: 1px solid #FEE2E2;
}

/* Test account cards */
.test-accounts {
  margin-top: 28px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.test-account-card {
  background: #F8FAFC;
  border: 1px solid #E2E8F0;
  border-radius: 8px;
  padding: 10px 14px;
}
.test-label {
  font-size: 11px;
  color: #94A3B8;
  font-weight: 500;
  letter-spacing: 0.04em;
  text-transform: uppercase;
  margin-bottom: 4px;
}
.test-cred {
  font-size: 13px;
  color: #334155;
  font-family: 'SF Mono', 'Fira Code', monospace;
}
</style>
