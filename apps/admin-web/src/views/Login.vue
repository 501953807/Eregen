<template>
  <div class="login-page">
    <div class="login-container">
      <div class="login-header">
        <h1>Eregen 管理后台</h1>
        <p class="subtitle">老年健康品牌平台 — 统一管理入口</p>
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
        label-width="80px"
        size="large"
        class="login-form"
      >
        <el-form-item label="邮箱" prop="email">
          <el-input v-model="emailForm.email" placeholder="请输入邮箱" />
        </el-form-item>
        <el-form-item label="密码" prop="password">
          <el-input
            v-model="emailForm.password"
            type="password"
            placeholder="请输入密码"
            show-password
          />
        </el-form-item>
        <el-form-item>
          <el-button
            type="primary"
            @click="submitEmailLogin"
            :loading="authStore.loading"
            style="width: 100%"
          >{{ authStore.loading ? '登录中...' : '登录' }}</el-button>
        </el-form-item>
        <div v-if="authStore.error" class="error-message">{{ authStore.error }}</div>
      </el-form>

      <!-- Phone / OTP Form -->
      <el-form
        v-show="activeTab === 'phone'"
        ref="phoneFormEl"
        :model="phoneForm"
        :rules="phoneRules"
        label-width="80px"
        size="large"
        class="login-form"
      >
        <el-form-item label="手机号" prop="phone">
          <el-input v-model="phoneForm.phone" placeholder="请输入手机号" type="tel" />
        </el-form-item>
        <el-form-item label="验证码" prop="otp">
          <el-row :gutter="12">
            <el-col :span="15">
              <el-input
                v-model="phoneForm.otp"
                placeholder="6位验证码"
                maxlength="6"
                type="digit"
              />
            </el-col>
            <el-col :span="9">
              <el-button
                type="primary"
                plain
                @click="sendOtp"
                :disabled="countdown > 0 || authStore.loading"
                style="width: 100%;"
              >{{ countdown > 0 ? countdown + 's' : '发送验证码' }}</el-button>
            </el-col>
          </el-row>
        </el-form-item>
        <el-form-item>
          <el-button
            type="primary"
            @click="submitPhoneLogin"
            :loading="authStore.loading"
            style="width: 100%"
          >{{ authStore.loading ? '登录中...' : '登录' }}</el-button>
        </el-form-item>
        <div v-if="authStore.error" class="error-message">{{ authStore.error }}</div>
      </el-form>

      <div class="hint">
        测试账号（邮箱登录）：<br />
        账号：admin@eregen.com / 密码：Admin@123<br /><br />
        测试账号（手机验证码）：<br />
        手机：13800000002 / 验证码：123456
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import type { FormRules } from 'element-plus'
import { ElMessage } from 'element-plus'
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
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  background: linear-gradient(135deg, #f5f7fa 0%, #c3cfe2 100%);
  padding: 20px;
}
.login-container {
  width: 100%;
  max-width: 440px;
  padding: 40px;
  background: #fff;
  border-radius: 12px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.08);
}
.login-header {
  text-align: center;
  margin-bottom: 28px;
}
.login-header h1 {
  color: #1a1a1a;
  font-size: 26px;
  font-weight: 600;
  margin: 0 0 6px;
}
.subtitle {
  color: #666;
  font-size: 13px;
  margin: 0;
}
.login-tabs {
  display: flex;
  border-bottom: 2px solid #ebeef5;
  margin-bottom: 24px;
}
.tab-item {
  flex: 1;
  text-align: center;
  padding: 10px 0;
  font-size: 15px;
  color: #909399;
  cursor: pointer;
  transition: all 0.2s;
  border-bottom: 2px solid transparent;
  margin-bottom: -2px;
}
.tab-item.active {
  color: #F59E0B;
  border-bottom-color: #F59E0B;
  font-weight: 600;
}
.login-form {
  margin-bottom: 8px;
}
.el-form-item {
  margin-bottom: 18px;
}
.error-message {
  color: #f56c6c;
  margin-top: 8px;
  font-size: 13px;
  text-align: center;
  padding: 6px 12px;
  background: #fef0f0;
  border-radius: 4px;
}
.hint {
  text-align: center;
  font-size: 12px;
  color: #999;
  margin-top: 20px;
  line-height: 1.8;
}
.hint code {
  background: #f5f7fa;
  padding: 1px 5px;
  border-radius: 3px;
  font-family: monospace;
  font-size: 12px;
  color: #333;
}
</style>
