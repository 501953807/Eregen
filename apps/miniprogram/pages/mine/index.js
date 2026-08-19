const app = getApp()
const { ApiClient } = require('../../utils/api')

Page({
  data: {
    elderlyList: [],
    currentElderlyId: 0,
    settings: { soundEnabled: true, vibrationEnabled: true, darkMode: false },
  },

  onLoad() { this.loadElderlyList() },

  async loadElderlyList() {
    try {
      const api = new ApiClient()
      const res = await api.get('/api/v1/admin/elderly?page_size=20')
      const list = (res?.data || []).map((e, i) => ({
        id: e.id,
        name: e.name,
        avatar: i % 2 === 0 ? '👴' : '👵',
        online: true,
      }))
      this.setData({ elderlyList: list.length > 0 ? list : app.globalData.elderlyList })
    } catch (e) {
      this.setData({ elderlyList: app.globalData.elderlyList })
    }
  },

  switchElderly(e) {
    this.setData({ currentElderlyId: e.currentTarget.dataset.index })
    wx.switchTab({ url: '/pages/home/index' })
  },

  addElderly() {
    wx.showModal({
      title: '添加老人',
      content: '请输入老人的身份信息',
      success: (res) => { if (res.confirm) wx.navigateTo({ url: '/pages/add-elderly/index' }) },
    })
  },

  toggleSetting(e) {
    const key = e.currentTarget.dataset.key
    const val = !this.data.settings[key]
    this.setData({ [`settings.${key}`]: val })
    wx.setStorageSync(`setting_${key}`, val)
  },

  clearCache() {
    wx.showModal({
      title: '清除缓存',
      content: '确定要清除本地缓存吗？',
      success: (res) => { if (res.confirm) { wx.clearStorageSync(); wx.showToast({ title: '已清除', icon: 'success' }) } },
    })
  },

  logout() {
    wx.showModal({
      title: '退出登录',
      content: '确定要退出登录吗？',
      success: (res) => { if (res.confirm) { wx.removeStorageSync('token'); wx.reLaunch({ url: '/pages/login/index' }) } },
    })
  },
})
