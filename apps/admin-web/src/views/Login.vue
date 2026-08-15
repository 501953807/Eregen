<template>
  <div class="login-page">
    <!-- Left decorative panel — Hope UI blue-purple gradient -->
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
            <svg width="36" height="36" viewBox="0 0 32 32" fill="none">
              <rect width="32" height="32" rx="8" fill="rgba(255,255,255,0.15)"/>
              <path d="M16 8L10 14v6l6 3 6-3v-6L16 8z" fill="rgba(255,255,255,0.9)" stroke="rgba(255,255,255,0.5)" stroke-width="0.5"/>
              <circle cx="16" cy="17" r="2" fill="rgba(92,141,115,0.9)"/>
            </svg>
          </div>
          <div>
            <div class="logo-text">Eregen</div>
            <div class="logo-sub">颐贞 · 康养管理平台</div>
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
          <span>© 2026 Eregen · 颐贞智能健康平台</span>
        </div>
      </div>
    </div>

    <!-- Right login form panel — white, Hope UI style -->
    <div class="login-form-panel">
      <div class="login-form-wrapper" :class="{ animating: visible }">

        <!-- Header -->
        <div class="login-header">
          <div class="login-header-top">
            <div class="login-header-logo">
              <svg width="18" height="18" viewBox="0 0 32 32" fill="none">
                <rect width="32" height="32" rx="8" fill="rgba(255,255,255,0.2)"/>
                <path d="M16 8L10 14v6l6 3 6-3v-6L16 8z" fill="rgba(255,255,255,0.9)"/>
                <circle cx="16" cy="17" r="2" fill="rgba(92,141,115,0.9)"/>
              </svg>
            </div>
            <span class="login-header-brand"><span>颐贞</span> · 管理后台</span>
          </div>
          <h1 class="login-title">欢迎登录</h1>
          <p class="login-subtitle">请输入您的账号信息以继续</p>
        </div>

        <!-- Tabs — pill container, Hope UI style -->
        <div class="login-tabs">
          <button
            class="login-tab"
            :class="{ active: activeTab === 'email' }"
            @click="activeTab = 'email'"
          >账号登录</button>
          <button
            class="login-tab"
            :class="{ active: activeTab === 'phone' }"
            @click="activeTab = 'phone'"
          >手机验证码</button>
        </div>

        <!-- Email / Password Form -->
        <form v-show="activeTab === 'email'" class="login-form" id="email-form" @submit.prevent="submitEmailLogin">
          <div class="form-group">
            <label class="form-label" for="email">邮箱地址</label>
            <div class="input-wrapper">
              <svg class="input-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <path d="M4 4h16c1.1 0 2 .9 2 2v12c0 1.1-.9 2-2 2H4c-1.1 0-2-.9-2-2V6c0-1.1.9-2 2-2z"/><polyline points="22,6 12,13 2,6"/>
              </svg>
              <input
                class="input"
                type="email"
                id="email"
                placeholder="请输入邮箱地址"
                v-model="emailForm.email"
                :class="{ 'input-error': emailError }"
              />
            </div>
            <div v-if="emailError" class="field-error">{{ emailError }}</div>
          </div>

          <div class="form-group">
            <label class="form-label" for="password">密码</label>
            <div class="input-wrapper">
              <svg class="input-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <rect x="3" y="11" width="18" height="11" rx="2" ry="2"/><path d="M7 11V7a5 5 0 0 1 10 0v4"/>
              </svg>
              <input
                class="input"
                :type="showPassword ? 'text' : 'password'"
                id="password"
                placeholder="请输入密码"
                v-model="emailForm.password"
                :class="{ 'input-error': passwordError }"
                @keyup.enter="submitEmailLogin"
              />
              <button type="button" class="pw-toggle" @click="showPassword = !showPassword">
                <svg v-if="showPassword" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19m-6.72-1.07a3 3 0 1 1-4.24-4.24"/><line x1="1" y1="1" x2="23" y2="23"/></svg>
                <svg v-else width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/><circle cx="12" cy="12" r="3"/></svg>
              </button>
            </div>
            <div v-if="passwordError" class="field-error">{{ passwordError }}</div>
          </div>

          <div class="form-extras">
            <label class="checkbox-wrapper">
              <input type="checkbox" v-model="rememberMe" />
              <span>记住登录状态</span>
            </label>
            <a class="forgot-link" href="#">忘记密码？</a>
          </div>

          <button type="submit" class="btn-login" :disabled="authStore.loading">
            <svg v-if="!authStore.loading" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M15 3h4a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2h-4"/><polyline points="10 17 15 12 10 7"/><line x1="15" y1="12" x2="3" y2="12"/>
            </svg>
            {{ authStore.loading ? '登录中...' : '登 录' }}
          </button>

          <div v-if="authStore.error" class="error-message">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <circle cx="12" cy="12" r="10"/><line x1="15" y1="9" x2="9" y2="15"/><line x1="9" y1="9" x2="15" y2="15"/>
            </svg>
            <span>{{ authStore.error }}</span>
          </div>
        </form>

        <!-- Phone / OTP Form -->
        <form v-show="activeTab === 'phone'" class="login-form" id="phone-form" @submit.prevent="submitPhoneLogin">
          <div class="form-group">
            <label class="form-label" for="phone">手机号</label>
            <div class="input-wrapper">
              <svg class="input-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <rect x="5" y="2" width="14" height="20" rx="2" ry="2"/><line x1="12" y1="18" x2="12.01" y2="18"/>
              </svg>
              <input
                class="input"
                type="tel"
                id="phone"
                placeholder="请输入手机号"
                v-model="phoneForm.phone"
                :class="{ 'input-error': phoneError }"
                maxlength="11"
              />
            </div>
            <div v-if="phoneError" class="field-error">{{ phoneError }}</div>
          </div>

          <div class="form-group">
            <label class="form-label" for="otp">验证码</label>
            <div class="otp-row">
              <div class="input-wrapper" style="flex:1;">
                <svg class="input-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <rect x="3" y="11" width="18" height="11" rx="2" ry="2"/><path d="M7 11V7a5 5 0 0 1 10 0v4"/>
                </svg>
                <input
                  class="input"
                  type="text"
                  id="otp"
                  placeholder="6位数字验证码"
                  v-model="phoneForm.otp"
                  :class="{ 'input-error': otpError }"
                  maxlength="6"
                />
              </div>
              <button
                type="button"
                class="btn-send-otp"
                :disabled="countdown > 0 || authStore.loading"
                @click="sendOtp"
              >{{ countdown > 0 ? countdown + 's' : '发送验证码' }}</button>
            </div>
            <div v-if="otpError" class="field-error">{{ otpError }}</div>
          </div>

          <button type="submit" class="btn-login" :disabled="authStore.loading">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M15 3h4a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2h-4"/><polyline points="10 17 15 12 10 7"/><line x1="15" y1="12" x2="3" y2="12"/>
            </svg>
            {{ authStore.loading ? '登录中...' : '登 录' }}
          </button>

          <div v-if="authStore.error" class="error-message">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <circle cx="12" cy="12" r="10"/><line x1="15" y1="9" x2="9" y2="15"/><line x1="9" y1="9" x2="15" y2="15"/>
            </svg>
            <span>{{ authStore.error }}</span>
          </div>
        </form>

        <!-- Divider + hint -->
        <div class="login-divider">测试账号</div>
        <div class="login-hint">
          <strong>邮箱登录</strong>：账号 <code>admin@eregen.com</code> · 密码 <code>Admin@123</code><br />
          <strong>手机登录</strong>：手机 <code>13800000002</code> · 验证码 <code>123456</code>
        </div>

        <div class="login-footer">© 2026 Eregen · 颐贞智能健康平台</div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { ElMessage } from 'element-plus'

const authStore = useAuthStore()
const route = useRoute()
const router = useRouter()

const activeTab = ref<'email' | 'phone'>('email')
const visible = ref(false)
const showPassword = ref(false)
const rememberMe = ref(true)

// --- Email / Password ---
const emailForm = ref({ email: '', password: '' })
const emailError = ref('')
const passwordError = ref('')

function validateEmail(): boolean {
  emailError.value = ''
  passwordError.value = ''
  if (!emailForm.value.email) { emailError.value = '请输入邮箱'; return false }
  if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(emailForm.value.email)) { emailError.value = '邮箱格式不正确'; return false }
  if (!emailForm.value.password) { passwordError.value = '请输入密码'; return false }
  if (emailForm.value.password.length < 6) { passwordError.value = '密码至少6位'; return false }
  return true
}

async function submitEmailLogin() {
  if (!validateEmail()) return
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
const phoneForm = ref({ phone: '', otp: '' })
const phoneError = ref('')
const otpError = ref('')
const countdown = ref(0)
const timerRef = ref<number | null>(null)

function validatePhone(): boolean {
  phoneError.value = ''
  otpError.value = ''
  if (!phoneForm.value.phone) { phoneError.value = '请输入手机号'; return false }
  if (!/^(\\+86|86)?1[3-9]\\d{9}$/.test(phoneForm.value.phone.replace(/\s/g, ''))) { phoneError.value = '请输入正确的中国大陆手机号'; return false }
  if (!phoneForm.value.otp) { otpError.value = '请输入验证码'; return false }
  if (!/^\\d{6}$/.test(phoneForm.value.otp)) { otpError.value = '验证码应为6位数字'; return false }
  return true
}

async function submitPhoneLogin() {
  if (!validatePhone()) return
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

onMounted(() => {
  requestAnimationFrame(() => { visible.value = true })
  if (authStore.checkLoggedIn()) {
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
  background: var(--hope-bg);
  overflow: hidden;
}

/* ==================== LEFT DECORATIVE PANEL ==================== */
.login-brand {
  flex: 1;
  min-height: 100vh;
  background: linear-gradient(135deg, #3a57e8 0%, #6f42c1 100%);
  position: relative;
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  padding: 60px;
  overflow: hidden;
}
.login-brand::before {
  content: '';
  position: absolute;
  inset: 0;
  background:
    radial-gradient(ellipse at 20% 20%, rgba(255,255,255,0.1) 0%, transparent 50%),
    radial-gradient(ellipse at 80% 80%, rgba(111,66,193,0.2) 0%, transparent 50%);
}
.login-brand::after {
  content: '';
  position: absolute;
  inset: 0;
  background-image: radial-gradient(circle at 1px 1px, rgba(255,255,255,0.04) 1px, transparent 0);
  background-size: 24px 24px;
}

.brand-bg-shapes {
  position: absolute;
  inset: 0;
  overflow: hidden;
  pointer-events: none;
}
.shape {
  position: absolute;
  border-radius: 50%;
  opacity: 0.1;
  background: #FFFFFF;
}
.shape-1 { width: 300px; height: 300px; top: -60px; right: -60px; animation: float 8s ease-in-out infinite; }
.shape-2 { width: 180px; height: 180px; bottom: 120px; left: -40px; animation: float 10s ease-in-out infinite reverse; }
.shape-3 { width: 100px; height: 100px; top: 45%; right: 30px; animation: float 6s ease-in-out infinite; }
.shape-4 { width: 60px; height: 60px; bottom: 200px; right: 80px; opacity: 0.15; animation: float 7s ease-in-out infinite reverse; }
@keyframes float {
  0%, 100% { transform: translateY(0); }
  50% { transform: translateY(-12px); }
}

.brand-content {
  position: relative;
  z-index: 1;
  text-align: center;
  max-width: 400px;
}

/* Logo */
.brand-logo {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
  margin-bottom: 32px;
}
.logo-icon {
  width: 48px;
  height: 48px;
  display: flex;
  align-items: center;
  justify-content: center;
}
.logo-text {
  font-size: 26px;
  font-weight: 800;
  color: #FFFFFF;
  letter-spacing: -0.02em;
  line-height: 1.1;
}
.logo-sub {
  font-size: 12px;
  color: rgba(255,255,255,0.55);
  letter-spacing: 0.08em;
  margin-top: 2px;
}

/* Feature list */
.brand-features {
  display: flex;
  flex-direction: column;
  gap: 12px;
  text-align: left;
}
.feature-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  background: rgba(255,255,255,0.08);
  border: 1px solid rgba(255,255,255,0.12);
  border-radius: var(--hope-radius-md);
  backdrop-filter: blur(8px);
  transition: background 0.2s ease;
}
.feature-item:hover {
  background: rgba(255,255,255,0.12);
}
.feature-dot {
  width: 32px;
  height: 32px;
  min-width: 32px;
  background: rgba(255,255,255,0.2);
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
}
.feature-dot::after {
  content: '';
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: rgba(255,255,255,0.85);
}
.feature-title {
  font-size: 13px;
  font-weight: 600;
  color: rgba(255,255,255,0.92);
  margin-bottom: 2px;
}
.feature-desc {
  font-size: 12px;
  color: rgba(255,255,255,0.55);
  line-height: 1.5;
}

.brand-footer {
  margin-top: 48px;
  font-size: 11px;
  color: rgba(255,255,255,0.4);
  letter-spacing: 0.02em;
}

/* ==================== RIGHT FORM PANEL ==================== */
.login-form-panel {
  flex: 0 0 520px;
  background: var(--hope-surface);
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  padding: 60px 48px;
  position: relative;
  box-shadow: -8px 0 32px rgba(17,38,146,0.06);
}
.login-form-panel::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  width: 3px;
  height: 100%;
  background: var(--hope-primary-gradient);
  background-size: 100% 200%;
  animation: gradientShift 4s ease infinite;
}

.login-form-wrapper {
  width: 100%;
  max-width: 380px;
  opacity: 0;
  transform: translateY(12px);
  transition: opacity 0.4s ease, transform 0.4s ease;
}
.login-form-wrapper.animating {
  opacity: 1;
  transform: translateY(0);
}

/* ==================== LOGIN HEADER ==================== */
.login-header {
  margin-bottom: 32px;
}
.login-header-top {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 20px;
}
.login-header-logo {
  width: 36px;
  height: 36px;
  background: var(--hope-primary-gradient);
  border-radius: var(--hope-radius-sm);
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 4px 12px rgba(58,87,232,0.25);
}
.login-header-brand {
  font-size: 17px;
  font-weight: 700;
  color: var(--hope-text);
  letter-spacing: -0.01em;
}
.login-header-brand span {
  background: var(--hope-primary-gradient);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}
.login-title {
  font-size: 24px;
  font-weight: 800;
  color: var(--hope-text);
  letter-spacing: -0.025em;
  margin-bottom: 6px;
}
.login-subtitle {
  font-size: 13px;
  color: var(--hope-text-muted);
}

/* ==================== TABS ==================== */
.login-tabs {
  display: flex;
  gap: 4px;
  background: var(--hope-primary-lighter);
  border-radius: var(--hope-radius-pill);
  padding: 4px;
  border: 1px solid var(--hope-border);
  margin-bottom: 28px;
  box-shadow: inset 0 1px 2px rgba(0,0,0,0.03);
}
.login-tab {
  flex: 1;
  padding: 9px 16px;
  border-radius: var(--hope-radius-pill);
  font-size: 13px;
  font-weight: 600;
  color: var(--hope-text-secondary);
  cursor: pointer;
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
  border: none;
  background: none;
  font-family: inherit;
  text-align: center;
}
.login-tab:hover {
  color: var(--hope-primary-hover);
  background: var(--hope-surface);
}
.login-tab.active {
  background: var(--hope-primary-gradient);
  color: #FFFFFF;
  box-shadow: var(--hope-shadow-active);
}

/* ==================== FORM ==================== */
.login-form {
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.form-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
}
.form-label {
  font-size: 12.5px;
  font-weight: 600;
  color: var(--hope-text-secondary);
  letter-spacing: 0.01em;
}
.input-wrapper {
  position: relative;
  display: flex;
  align-items: center;
}
.input-icon {
  position: absolute;
  left: 12px;
  width: 18px;
  height: 18px;
  color: var(--hope-text-muted);
  transition: color 0.2s ease;
  pointer-events: none;
}
.input-wrapper:focus-within .input-icon {
  color: var(--hope-primary);
}
.input {
  height: 42px;
  padding: 0 40px;
  border: 1px solid var(--hope-border);
  border-radius: var(--hope-radius-sm);
  font-size: 14px;
  font-family: inherit;
  color: var(--hope-text);
  background: var(--hope-surface);
  box-shadow: inset 0 1px 2px rgba(0,0,0,0.04);
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
  outline: none;
  width: 100%;
  box-sizing: border-box;
}
.input:hover { border-color: var(--hope-border-strong); }
.input:focus {
  border-color: var(--hope-primary);
  box-shadow: var(--hope-shadow-input-focus);
}
.input::placeholder { color: var(--hope-text-muted); }
.input-error {
  border-color: var(--hope-danger) !important;
  box-shadow: 0 0 0 3px rgba(192,50,33,0.1) !important;
}
.pw-toggle {
  position: absolute;
  right: 12px;
  background: none;
  border: none;
  cursor: pointer;
  padding: 4px;
  color: var(--hope-text-muted);
  display: flex;
  align-items: center;
  transition: color 0.2s;
}
.pw-toggle:hover { color: var(--hope-primary); }

/* OTP row */
.otp-row {
  display: flex;
  gap: 10px;
}
.otp-row .input { flex: 1; padding-right: 12px; }
.btn-send-otp {
  height: 42px;
  padding: 0 16px;
  border-radius: var(--hope-radius-sm);
  border: 1px solid var(--hope-border);
  background: var(--hope-surface);
  color: var(--hope-primary);
  font-size: 13px;
  font-weight: 600;
  font-family: inherit;
  cursor: pointer;
  transition: all 0.2s;
  white-space: nowrap;
  flex-shrink: 0;
}
.btn-send-otp:hover {
  background: var(--hope-primary-lighter);
  border-color: var(--hope-primary);
}
.btn-send-otp:disabled {
  color: var(--hope-text-muted);
  cursor: not-allowed;
  background: var(--hope-surface-light);
  border-color: var(--hope-border);
}

/* Field error */
.field-error {
  font-size: 12px;
  color: var(--hope-danger);
  margin-top: 2px;
}

/* Extras row */
.form-extras {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 2px;
}
.checkbox-wrapper {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  font-size: 13px;
  color: var(--hope-text-secondary);
}
.checkbox-wrapper input[type="checkbox"] {
  width: 16px;
  height: 16px;
  accent-color: var(--hope-primary);
  cursor: pointer;
}
.forgot-link {
  font-size: 13px;
  color: var(--hope-primary);
  text-decoration: none;
  font-weight: 500;
  transition: color 0.2s;
}
.forgot-link:hover { color: var(--hope-primary-hover); text-decoration: underline; }

/* ==================== LOGIN BUTTON ==================== */
.btn-login {
  width: 100%;
  height: 44px;
  border-radius: var(--hope-radius-sm);
  border: none;
  background: var(--hope-primary-gradient);
  color: #FFFFFF;
  font-size: 15px;
  font-weight: 700;
  font-family: inherit;
  cursor: pointer;
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
  box-shadow: var(--hope-shadow-primary);
  margin-top: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
}
.btn-login:hover:not(:disabled) {
  background: var(--hope-primary-gradient-hover);
  box-shadow: var(--hope-shadow-active);
  transform: translateY(-1px);
}
.btn-login:active:not(:disabled) { transform: translateY(0); }
.btn-login:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}
.btn-login svg {
  width: 18px;
  height: 18px;
}

/* ==================== ERROR MESSAGE ==================== */
.error-message {
  padding: 10px 14px;
  border-radius: var(--hope-radius-md);
  background: var(--hope-danger-light);
  border: 1px solid var(--hope-danger-light);
  color: var(--hope-danger);
  font-size: 13px;
  text-align: center;
  display: flex;
  align-items: center;
  gap: 8px;
}
.error-message svg {
  width: 16px;
  height: 16px;
  flex-shrink: 0;
}

/* ==================== DIVIDER ==================== */
.login-divider {
  display: flex;
  align-items: center;
  gap: 12px;
  margin: 24px 0;
  color: var(--hope-text-muted);
  font-size: 12px;
}
.login-divider::before,
.login-divider::after {
  content: '';
  flex: 1;
  height: 1px;
  background: var(--hope-border);
}

/* ==================== HINT ==================== */
.login-hint {
  text-align: center;
  font-size: 12px;
  color: var(--hope-text-muted);
  line-height: 1.8;
  padding: 16px;
  background: var(--hope-surface-light);
  border-radius: var(--hope-radius-md);
  border: 1px solid var(--hope-border);
}
.login-hint strong {
  color: var(--hope-text-secondary);
  font-weight: 600;
}
.login-hint code {
  background: var(--hope-surface);
  padding: 1px 6px;
  border-radius: var(--hope-radius-sm);
  font-family: 'SF Mono', 'Consolas', monospace;
  font-size: 11px;
  color: var(--hope-primary);
  border: 1px solid var(--hope-border);
}

/* ==================== FOOTER ==================== */
.login-footer {
  margin-top: 32px;
  text-align: center;
  font-size: 12px;
  color: var(--hope-text-muted);
}

@keyframes gradientShift {
  0% { background-position: 0% 50%; }
  50% { background-position: 100% 50%; }
  100% { background-position: 0% 50%; }
}

/* ==================== RESPONSIVE ==================== */
@media (max-width: 900px) {
  .login-brand { display: none; }
  .login-form-panel {
    flex: 1;
    box-shadow: none;
  }
  .login-form-panel::before { width: 100%; height: 3px; }
}
</style>
