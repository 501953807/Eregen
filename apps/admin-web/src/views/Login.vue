<template>
  <div class="login-page">
    <div class="login-container">
      <div class="login-header">
        <h1>Eregen 管理后台</h1>
        <p class="subtitle">老年健康品牌平台 — 统一管理入口</p>
      </div>

      <div class="login-tabs">
        <el-tab-bar v-model="activeTab" class="tab-switcher">
          <el-tab-item label="账号登录" name="email" />
          <el-tab-item label="手机验证码" name="phone" />
        </el-tab-bar>
      </div>

      <el-form ref="emailFormEl" :model="emailForm" :rules="emailRules" label-width="120px" size="large" class="login-form" :class="{ 'form-active': activeTab === 'email' }">
        <el-form-item label="邮箱" prop="email"><el-input v-model="emailForm.email" placeholder="请输入注册邮箱" type="email" /></el-form-item>
        <el-form-item label="密码" prop="password"><el-input v-model="emailForm.password" type="password" placeholder="请输入密码" show-password /></el-form-item>
        <el-form-item><el-button type="primary" @click="submitEmailLogin" :loading="authStore.loading" style="width: 100%">{{ authStore.loading ? '登录中...' : '登录' }}</el-button></el-form-item>
        <div v-if="authStore.error" class="error-message">{{ authStore.error }}</div>
      </el-form>

      <el-form ref="phoneFormEl" :model="phoneForm" :rules="phoneRules" label-width="120px" size="large" class="login-form" :class="{ 'form-active': activeTab === 'phone' }">
        <el-form-item label="手机号" prop="phone"><el-input v-model="phoneForm.phone" placeholder="请输入绑定手机号 (+86...)" type="tel" /></el-form-item>
        <el-row :gutter="16"><el-col :span="14"><el-form-item label="验证码" prop="otp"><el-input v-model="phoneForm.otp" placeholder="6位验证码" maxlength="6" type="digit" /></el-form-item></el-col><el-col :span="10"><el-button type="default" @click="sendOtp" :disabled="countdown > 0 || authStore.loading" class="otp-button">{{ countdown > 0 ? countdown + 's' : '发送验证码' }}</el-button></el-col></el-row>
        <el-form-item><el-button type="primary" @click="submitPhoneLogin" :loading="authStore.loading" style="width: 100%">{{ authStore.loading ? '登录中...' : '登录' }}</el-button></el-form-item>
        <div v-if="authStore.error" class="error-message">{{ authStore.error }}</div>
      </el-form>

      <div class="hint">
        默认测试账号：<br>
        📧 admin@eregen.com / Admin@123 （管理员）<br>
        📱 13800000002 / 123456 （家属/操作员）
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import type { FormRules } from 'element-plus'
import { ElMessage, ElTabBar, ElTabItem } from 'element-plus'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()
const route = useRoute()
const router = useRouter()

const activeTab = ref<'email' | 'phone'>('email')

const emailFormEl = ref<any>(null)
const emailForm = ref({ email: '', password: '' })
const emailRules = computed<FormRules>(() => ({
  email: [{ required: true, message: '请输入邮箱', trigger: 'blur' }, { pattern: /^[^\s@]+@[^\s@]+\.[^\s@]+$/, message: '邮箱格式不正确', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }, { min: 6, message: '密码至少6位', trigger: 'blur' }]
}))
const phoneFormEl = ref<any>(null)
const phoneForm = ref({ phone: '', otp: '' })
const phoneRules = computed<FormRules>(() => ({
  phone: [{ required: true, message: '请输入手机号', trigger: 'blur' }, { pattern: /^(\+86|86)?1[3-9]\d{9}$/, message: '请输入正确的中国大陆手机号', trigger: 'blur' }],
  otp: [{ required: true, message: '请输入验证码', trigger: 'blur' }, { pattern: /^\d{6}$/, message: '验证码应为6位数字', trigger: 'blur' }]
}))
const countdown = ref(0)
const timerRef = ref<number | null>(null)

const sendOtp = async () => {
  if (!phoneForm.value.phone) { ElMessage.error('请输入手机号'); return }
  ElMessage.success('验证码已发送，测试用：123456')
  countdown.value = 60
  timerRef.value = window.setInterval(() => { countdown.value--; if (countdown.value <= 0) { clearInterval(timerRef.value!); countdown.value = 0 } }, 1000)
}

const submitEmailLogin = async () => {
  if (!emailFormEl.value) return
  const valid = await emailFormEl.value.validate()
  if (!valid) return
  try { await authStore.login({ method: 'email', credential: emailForm.value.email, secret: emailForm.value.password }) }
  catch (err: any) { authStore.error.value = err.response?.data?.msg || '登录失败，请检查账户信息' }
}

const submitPhoneLogin = async () => {
  if (!phoneFormEl.value) return
  const valid = await phoneFormEl.value.validate()
  if (!valid) return
  try { await authStore.login({ method: 'phone', credential: phoneForm.value.phone, secret: phoneForm.value.otp }) }
  catch (err: any) { authStore.error.value = err.response?.data?.msg || '登录失败，请检查手机号和验证码' }
}

onMounted(() => { if (authStore.isLoggedIn) { const to = route.query.redirect || '/dashboard'; router.push({ path: to as string }) } })
onBeforeUnmount(() => { if (timerRef.value) clearInterval(timerRef.value!) })
</script>

<style scoped>
.login-page { display: flex; justify-content: center; align-items: center; min-height: 100vh; background: linear-gradient(135deg, #f5f7fa 0%, #c3cfe2 100%); padding: 20px; }
.login-container { width: 100%; max-width: 480px; padding: 48px; background: white; border-radius: 12px; box-shadow: 0 8px 32px rgba(0,0,0,0.08); }
.login-header { text-align: center; margin-bottom: 32px; }
.login-header h1 { color: #1a1a1a; font-size: 28px; font-weight: 600; margin-bottom: 8px; }
.subtitle { color: #666; font-size: 14px; margin: 0; }
.tab-switcher { margin-bottom: 32px; --el-tabbar-border-color: transparent; --el-tab-background: transparent; --el-tab-text-color: #999; --el-tab-active-text-color: #F59E0B; --el-tab-hover-color: transparent; }
.login-form { display: none; opacity: 0; transform: translateY(10px); transition: all 0.3s ease; }
.login-form.form-active { display: block; opacity: 1; transform: translateY(0); }
.el-form-item { margin-bottom: 20px; }
.otp-button { width: 100%; margin-top: 12px; font-size: 14px; color: #F59E0B; border-color: #F59E0B; }
.otp-button:disabled { color: #ccc; border-color: #ccc; cursor: not-allowed; }
.hint { text-align: center; font-size: 13px; color: #999; margin-top: 24px; line-height: 1.6; }
.hint code { background: #f5f7fa; padding: 2px 6px; border-radius: 4px; font-family: monospace; font-size: 12px; color: #333; }
.error-message { color: #f56c6c; margin-top: 16px; font-size: 13px; text-align: center; padding: 8px; background: #fff5f5; border-radius: 4px; }
:root { --brand-primary: #F59E0B; --brand-primary-dark: #D97706; }
.el-button--primary { background: var(--brand-primary) !important; border-color: var(--brand-primary) !important; }
.el-button--primary:hover { background: var(--brand-primary-dark) !important; border-color: var(--brand-primary-dark) !important; }
.el-button--default { color: var(--brand-primary); border-color: var(--brand-primary); }
.el-button--default:hover { background: var(--brand-primary); color: white; }
@media (max-width: 480px) { .login-container { padding: 32px 24px; margin: 20px; } .login-header h1 { font-size: 24px; } }
</style>
