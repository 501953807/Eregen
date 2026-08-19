const { request } = require('../../utils/api')
const { set: storageSet, get: storageGet } = require('../../utils/storage')

Page({
  data: {
    phone: '',
    code: '',
    countdown: 0,
    loading: false,
  },

  onCountdownFinish() {
    this.setData({ countdown: 0 })
  },

  startCountdown() {
    let sec = 60
    this.setData({ countdown: sec })
    const timer = setInterval(() => {
      sec--
      if (sec <= 0) {
        clearInterval(timer)
        this.onCountdownFinish()
      } else {
        this.setData({ countdown: sec })
      }
    }, 1000)
  },

  onPhoneInput(e) {
    this.setData({ phone: e.detail.value })
  },

  onCodeInput(e) {
    this.setData({ code: e.detail.value })
  },

  sendCode() {
    const { phone } = this.data
    if (!/^1[3-9]\d{9}$/.test(phone)) {
      wx.showToast({ title: '请输入正确的手机号', icon: 'none' })
      return
    }
    if (this.data.countdown > 0) return
    this.setData({ loading: true })
    // SMS send endpoint not yet implemented; use mock verification
    setTimeout(() => {
      wx.showToast({ title: '验证码已发送（测试用：123456）', icon: 'success' })
      this.startCountdown()
      this.setData({ loading: false })
    }, 500)
  },

  login() {
    const { phone, code } = this.data
    if (!/^1[3-9]\d{9}$/.test(phone)) {
      wx.showToast({ title: '请输入正确的手机号', icon: 'none' })
      return
    }
    if (!code || code.length < 4) {
      wx.showToast({ title: '请输入验证码', icon: 'none' })
      return
    }
    this.setData({ loading: true })
    request('/api/v1/auth/login', { method: 'phone', credential: phone, secret: code }, 'POST')
      .then((res) => {
        const token = res?.token || (res?.data && res.data.token)
        const userInfo = res?.user || (res?.data && res.data.user)
        if (token) {
          storageSet('token', token)
          wx.setStorageSync('token', token)
          if (userInfo) {
            storageSet('user_info', userInfo)
            wx.setStorageSync('user_info', userInfo)
          }
          wx.showToast({ title: '登录成功', icon: 'success' })
          return request('/api/v1/admin/elderly', {}, 'GET')
            .then((res2) => {
              const elderlyList = Array.isArray(res2) ? res2 : (res2?.data || [])
              const hasElderly = elderlyList.length > 0
              return hasElderly ? '/pages/home/index' : '/pages/add-elderly/index'
            })
        }
        return '/pages/home/index'
      })
      .then((url) => {
        setTimeout(() => {
          wx.reLaunch({ url })
        }, 1500)
      })
      .catch((e) => {
        wx.showToast({ title: e.message || '登录失败', icon: 'none' })
      })
      .finally(() => {
        this.setData({ loading: false })
      })
  },

  onReady() {
    const savedPhone = storageGet('phone')
    if (savedPhone) {
      this.setData({ phone: savedPhone })
    }
  },
})
