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
    request('/auth/send-code', { phone }, 'POST')
      .then(() => {
        wx.showToast({ title: '验证码已发送', icon: 'success' })
        this.startCountdown()
      })
      .catch((e) => {
        wx.showToast({ title: e.message || '发送失败', icon: 'none' })
      })
      .finally(() => {
        this.setData({ loading: false })
      })
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
    request('/auth/phone-login', { phone, code }, 'POST')
      .then((res) => {
        if (res.token) {
          storageSet('token', res.token)
          wx.setStorageSync('token', res.token)
          wx.showToast({ title: '登录成功', icon: 'success' })
          // Check if user has elderly list
          return request('/elderly?owner_user_id=self', {}, 'GET')
            .then((res2) => {
              const hasElderly = Array.isArray(res2.data) && res2.data.length > 0
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
    // Pre-fill phone if stored
    const savedPhone = storageGet('phone')
    if (savedPhone) {
      this.setData({ phone: savedPhone })
    }
  },
})
