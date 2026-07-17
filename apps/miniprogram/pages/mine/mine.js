const api = require('../../utils/api')

Page({
  data: {
    userName: '管理员',
    userRole: '超级管理员',
    avatar: '👤',
    menuItems: [
      { title: '我的设备', icon: '📱', path: '/pages/home/index' },
      { title: '订阅管理', icon: '📋', path: '' },
      { title: '消息通知', icon: '🔔', path: '' },
      { title: '设置', icon: '⚙️', path: '' },
    ],
    settings: {
      pushEnabled: true,
      soundEnabled: true,
      vibrationEnabled: true,
      language: 'zh-CN',
    },
  },

  onLoad() {
    this.fetchUserProfile()
    this.fetchSettings()
  },

  async fetchUserProfile() {
    try {
      const token = wx.getStorageSync('token')
      if (!token) return
      const res = await api.request('/users/me', {}, 'GET')
      if (res.data || res.user) {
        const user = res.data || res.user
        this.setData({
          userName: user.name || '用户',
          userRole: user.role || '家属',
        })
      }
    } catch (e) {
      console.log('fetch profile failed:', e)
    }
  },

  async fetchSettings() {
    try {
      const token = wx.getStorageSync('token')
      if (!token) return
      const res = await api.request('/users/settings', {}, 'GET')
      if (res.data) {
        this.setData({ settings: res.data })
      }
    } catch (e) {
      console.log('fetch settings failed:', e)
    }
  },

  handleMenuTap(e) {
    const index = e.currentTarget.dataset.index
    const item = this.data.menuItems[index]
    if (item.path) {
      wx.navigateTo({ url: item.path })
    } else {
      switch (index) {
        case 1: // 订阅管理
          wx.showToast({ title: '订阅管理功能开发中', icon: 'none' })
          break
        case 2: // 消息通知
          wx.showToast({ title: '消息通知功能开发中', icon: 'none' })
          break
        case 3: // 设置
          this.showSettingsDialog()
          break
        default:
          wx.showToast({ title: '功能开发中', icon: 'none' })
      }
    }
  },

  showSettingsDialog() {
    const s = this.data.settings
    wx.showModal({
      title: '设置',
      content: `推送通知: ${s.pushEnabled ? '开启' : '关闭'}\n声音: ${s.soundEnabled ? '开启' : '关闭'}\n震动: ${s.vibrationEnabled ? '开启' : '关闭'}\n语言: ${s.language}`,
      showCancel: true,
      cancelText: '取消',
      confirmText: '保存',
      success: (res) => {
        if (res.confirm) {
          this.saveSettings()
        }
      },
    })
  },

  async saveSettings() {
    try {
      const token = wx.getStorageSync('token')
      if (!token) {
        wx.showToast({ title: '请先登录', icon: 'error' })
        return
      }
      await api.request('/users/settings', this.data.settings, 'PUT')
      wx.showToast({ title: '设置已保存', icon: 'success' })
    } catch (e) {
      wx.showToast({ title: '保存失败', icon: 'error' })
    }
  },

  onPushChange(e) {
    const s = { ...this.data.settings }
    s.pushEnabled = e.detail.value
    this.setData({ settings: s })
  },

  onSoundChange(e) {
    const s = { ...this.data.settings }
    s.soundEnabled = e.detail.value
    this.setData({ settings: s })
  },

  onVibrationChange(e) {
    const s = { ...this.data.settings }
    s.vibrationEnabled = e.detail.value
    this.setData({ settings: s })
  },

  onLogout() {
    wx.showModal({
      title: '确认退出',
      content: '确定要退出登录吗？',
      success: (res) => {
        if (res.confirm) {
          wx.removeStorageSync('token')
          wx.reLaunch({ url: '/pages/login/index' })
        }
      },
    })
  },
})
