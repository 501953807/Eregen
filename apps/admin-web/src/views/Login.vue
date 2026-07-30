<template>
  <div class="login-page">
    <div class="login-container">
      <h1>Eregen 管理后台</h1>
      <p class="subtitle">老年健康品牌平台 - 管理系统</p>

      <el-form
        ref="loginFormEl"
        :model="form"
        :rules="rules"
        label-width="120px"
        size="large"
      >
        <el-form-item label="用户名" prop="username">
          <el-input v-model="form.username" placeholder="请输入用户名" />
        </el-form-item>

        <el-form-item label="密码" prop="password">
          <el-input
            v-model="form.password"
            type="password"
            placeholder="请输入密码"
            show-password
          />
        </el-form-item>

        <el-form-item>
          <el-button
            type="primary"
            @click="submitLogin"
            :loading="authStore.loading"
            style="width: 100%"
          >
            登录
          </el-button>
        </el-form-item>

        <div v-if="authStore.error" class="error-message">
          {{ authStore.error }}
        </div>
      </el-form>

      <div class="hint">
        默认凭证：<code>admin / Admin@123</code>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import type { FormRules } from 'element-plus'
import { ElMessage } from 'element-plus'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()
const route = useRoute()
const router = useRouter()

const formRef = ref<any>(null)
const form = ref({
  username: '',
  password: ''
})

const rules = computed<FormRules>(() => ({
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' }
  ]
}))

const loginFormEl = ref(null)

async function submitLogin() {
  if (!loginFormEl.value) return

  const valid = await loginFormEl.value.validate()
  if (!valid) return

  try {
    await authStore.login(form.value.username, form.value.password)
    // After successful login, redirect to dashboard or previous route
    const to = route.query.redirect || '/dashboard'
    router.push(to as string)
  } catch (err: any) {
    console.error('Login error:', err)
  }
}

// On load, if already logged in, redirect to dashboard
onMounted(() => {
  if (authStore.isLoggedIn()) {
    const to = route.query.redirect || '/dashboard'
    router.push(to as string)
  }
})
</script>

<style scoped>
.login-page {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  background: #f5f7fa;
}

.login-container {
  width: 100%;
  max-width: 420px;
  padding: 40px;
  background: white;
  border-radius: 8px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.1);
}

.login-container h1 {
  text-align: center;
  color: #333;
  font-size: 28px;
  margin-bottom: 10px;
}

.subtitle {
  text-align: center;
  color: #666;
  font-size: 14px;
  margin-bottom: 30px;
}

.hint {
  text-align: center;
  font-size: 13px;
  color: #999;
  margin-top: 20px;
}

.hint code {
  background: #f0f0f0;
  padding: 2px 6px;
  border-radius: 3px;
  font-family: monospace;
}

.error-message {
  color: #f56c6c;
  margin-top: 15px;
  font-size: 14px;
}
</style>